package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"

	"log"
	"os"

	"strings"

	"./rewriter"

	"golang.org/x/tools/go/ast/astutil"
)

// TODO: comments are moved around by this script. Look at go/ast's CommentMap
// https://golang.org/pkg/go/ast/#CommentMap

// This script will emulate spacing that it sees in source code.
// For example, if it sees the spacing seen below:
//		b1 := bson.M{
//			"apple":    1,
//		}
//
// It will create the following new code:
//	b1 := bsonutil.NewD(
//		bsonutil.NewDocElem("apple", 1),
//	)

// The method by which the functions below insert newlines is a bit hacky,
// (by necessity). We insert the *ast.Ident "theremin" between arguments to functions, and
// replace those idents with newlines using sed within replace.sh.
// In the above example, the output of this go script would be:
// b1 := bsonutil.NewD(theremin, bsonutil.NewDocElem("a", 1), theremin)
// which would be ultimately converted to:
//	b1 := bsonutil.NewD(
//		bsonutil.NewDocElem("apple", 1),
//	)

// So, why are we doing this? There are no methods for adding newlines
// using the packages imported above.
// go/token has an AddLine method. However, this one can only be used
// to add a line to the very end of a file. go/printer will always print a CallExpr
// on a single line (https://golang.org/src/go/printer/nodes.go, line 150).

// Why "theremin"? "theremin" was chosen as the replacement string because
// it struck me as a word not likely to be found in the
// codebase. If "theremin" ever does come to live in the codebase,
// choose a new word!
var replace = ast.NewIdent("theremin")

type bsonRewriter struct {
	fs      *token.FileSet
	Changed bool
}

func countOpenBracketNewline(s string) int {
	return strings.Count(s, "{\n")
}

func countNewlineCloseBracket(s string) int {
	return strings.Count(s, "\n}")
}

func numberOfCommaNewlines(s string) int {
	return strings.Count(s, ",\n")
}

// shouldNewlines will look at the original bson.X{} call, and return a boolean
// for the the presence of a newline after the opening brace, before the closing
// brace, and between arguments.
func (rw *bsonRewriter) shouldNewlines(typedN ast.Expr) (beg, end, args bool) {
	// start out false
	// var beg, end, args bool
	buf := new(bytes.Buffer)

	if err := printer.Fprint(buf, rw.fs, typedN); err != nil {
		log.Fatal(err)
	}

	a := buf.String()

	if countOpenBracketNewline(a) > 0 {
		beg = true
	}
	if countNewlineCloseBracket(a) > 0 {
		end = true
	}

	if numberOfCommaNewlines(a) > 0 {
		args = true
	}
	return beg, end, args
}

// newCallExpr returns a new *ast.CallExpr of the form pkg.fxn(<args>)
func newCallExpr(pkg string, fxn string, args []ast.Expr) *ast.CallExpr {
	pkgIdent := ast.NewIdent(pkg)
	fxnIdent := ast.NewIdent(fxn)

	selExpr := &ast.SelectorExpr{X: pkgIdent, Sel: fxnIdent}
	callExpr := &ast.CallExpr{Fun: selExpr, Args: args}
	return callExpr
}

// handleBsonDocElem will convert a bson.DocElem{Name:<name>, Value:<value>} to bsonutil.NewDocElem(name, value)
// This function defaults to not adding newlines.
func (rw *bsonRewriter) handleBsonDocElem(typedN *ast.CompositeLit) (ast.Node, rewriter.Rewriter) {

	rw.Changed = true

	subArgs := []ast.Expr{}

	// iterate over elements of composite lit
	// e.g.Name: "a",  Value: 5
	for _, subElt := range typedN.Elts {

		kvExpr := subElt.(*ast.KeyValueExpr)
		value := rw.handleInterfaceArray(kvExpr.Value)
		subArgs = append(subArgs, value)
	}

	return newCallExpr("bsonutil", "NewDocElem", subArgs), rw
}

// handleInterfaceArray will check if val is an interface array, and if
// so, convert it to a bsonutil.NewArray().
// This function will always add newlines between items in bsonutil.NewArray.
func (rw *bsonRewriter) handleInterfaceArray(val ast.Expr) ast.Expr {
	// check that it is an interface array
	// if not, return val
	cl, ok := val.(*ast.CompositeLit)
	if !ok {
		return val
	}
	arr, ok := cl.Type.(*ast.ArrayType)
	if !ok {
		return val
	}

	_, ok = arr.Elt.(*ast.InterfaceType)
	if !ok {
		return val
	}

	args := []ast.Expr{}

	// newline after opening paren of NewArray
	args = append(args, replace)

	for _, item := range cl.Elts {
		// call this recursively to deal with interface arrays within interface arrays
		args = append(args, rw.handleInterfaceArray(item))

		// put newlines between items in array
		args = append(args, replace)
	}

	return newCallExpr("bsonutil", "NewArray", args)
}

// handleBsonM will convert "bson.M{<kv pairs>} into bsonutil.NewM(<DocElems>)"
func (rw *bsonRewriter) handleBsonM(typedN *ast.CompositeLit) (ast.Node, rewriter.Rewriter) {
	rw.Changed = true
	beg, end, commas := rw.shouldNewlines(typedN)

	args := []ast.Expr{}

	if beg {
		args = append(args, replace)
	}

	for _, item := range typedN.Elts {

		switch typedElt := item.(type) {

		case *ast.KeyValueExpr:
			// "k:v" -> "NewDocElem(k,v)"
			value := rw.handleInterfaceArray(typedElt.Value)
			kv := []ast.Expr{}
			kv = append(kv, typedElt.Key, value)

			args = append(args, newCallExpr("bsonutil", "NewDocElem", kv))

			if commas {
				args = append(args, replace)
			}

		default:
			//fmt.Printf("in handleBsonM, Bson.M Elts was not of type KeyValueExpr, type %T. \n", typedElt)
			//fmt.Println("Appending anyways...")
			args = append(args, item)

			if commas {
				args = append(args, replace)
			}

		}
	}

	if end {
		args = append(args, replace)
	}

	newNode := newCallExpr("bsonutil", "NewM", args)

	return newNode, rw
}

// handleBsonD will convert "bson.D{<DocElems>} into bsonutil.NewD(<DocElems>)"
func (rw *bsonRewriter) handleBsonD(typedN *ast.CompositeLit) (ast.Node, rewriter.Rewriter) {
	beg, end, commas := rw.shouldNewlines(typedN)
	rw.Changed = true
	args := []ast.Expr{}

	if beg {
		args = append(args, replace)
	}

	for _, item := range typedN.Elts {

		switch typedElt := item.(type) {
		// We know that its going to be two Elts, one "Name" and one "Value"
		case *ast.CompositeLit:
			subArgs := []ast.Expr{}
			for _, subElt := range typedElt.Elts {
				kvExpr := subElt.(*ast.KeyValueExpr)
				value := rw.handleInterfaceArray(kvExpr.Value)
				subArgs = append(subArgs, value)
			}

			args = append(args, newCallExpr("bsonutil", "NewDocElem", subArgs))
			if commas {
				args = append(args, replace)
			}

		default:

			//fmt.Printf("in handleBsonD, Bson.D Elts was not of type CompositeLit, type %T. \n", typedElt)
			//fmt.Println("Appending anyways...")
			args = append(args, item)

			if commas {
				args = append(args, replace)
			}
		}
	}

	if end {
		args = append(args, replace)
	}

	newNode := newCallExpr("bsonutil", "NewD", args)

	return newNode, rw
}

// handleBsonMArray will convert "[]bson.M{<bson.M's>} into bsonutil.NewMArray(<bson.M's>)"
func (rw *bsonRewriter) handleBsonMArray(typedN *ast.CompositeLit) (ast.Node, rewriter.Rewriter) {
	rw.Changed = true
	beg, end, commas := rw.shouldNewlines(typedN)
	args := []ast.Expr{}

	if beg {
		args = append(args, replace)
	}

	for _, item := range typedN.Elts {

		switch typedElt := item.(type) {

		case *ast.CompositeLit:
			bsonM, _ := rw.handleBsonM(typedElt)
			bsonMCast := bsonM.(ast.Expr)
			args = append(args, bsonMCast)

			if commas {
				args = append(args, replace)
			}

		case *ast.KeyValueExpr:
			fmt.Println("WARNING: Bson.M Array Elts contained naked k:v pair")

		default:
			//fmt.Printf("Bson.M Array Elts contained non CompositeLit, type %T \n", typedElt)
			//fmt.Println("Appending anyways...")
			args = append(args, item)

			if commas {
				args = append(args, replace)
			}

		}
	}

	if end {
		args = append(args, replace)
	}

	newNode := newCallExpr("bsonutil", "NewMArray", args)

	return newNode, rw
}

// handleBsonDArray will convert "[]bson.D{<bson.D's>} into bsonutil.NewDArray(<bson.D's>)"
func (rw *bsonRewriter) handleBsonDArray(typedN *ast.CompositeLit) (ast.Node, rewriter.Rewriter) {
	rw.Changed = true
	beg, end, commas := rw.shouldNewlines(typedN)
	args := []ast.Expr{}

	if beg {
		args = append(args, replace)
	}

	for _, item := range typedN.Elts {

		switch typedElt := item.(type) {
		// bson.D's
		case *ast.CompositeLit:
			bsonD, _ := rw.handleBsonD(typedElt)
			bsonDCast := bsonD.(ast.Expr)
			args = append(args, bsonDCast)
			if commas {
				args = append(args, replace)
			}

		// error cases
		case *ast.KeyValueExpr:
			fmt.Println("WARNING: Bson.D Array Elts contained naked k:v pair")

		default:
			//fmt.Printf("Bson.D Array Elts was not CompositeLit, but %T \n", typedElt)
			//fmt.Println("Appending anyways...")
			args = append(args, typedElt)
			if commas {
				args = append(args, replace)
			}

		}

	}
	if end {
		args = append(args, replace)
	}

	newNode := newCallExpr("bsonutil", "NewDArray", args)
	return newNode, rw
}

func (rw *bsonRewriter) Rewrite(node ast.Node) (ast.Node, rewriter.Rewriter) {
	newNode := node

	switch typedN := node.(type) {
	case *ast.CompositeLit:

		switch typedType := typedN.Type.(type) {
		case *ast.ArrayType:

			typedElt := typedType.Elt

			selExpr, ok := typedElt.(*ast.SelectorExpr)
			if !ok {
				break
			}

			if fmt.Sprintf("%v", selExpr) == "&{bson M}" {
				newNode, _ = rw.handleBsonMArray(typedN)
			}

			if fmt.Sprintf("%v", selExpr) == "&{bson D}" {
				newNode, _ = rw.handleBsonDArray(typedN)
			}

		}

		if fmt.Sprintf("%v", typedN.Type) == "&{bson M}" {
			newNode, _ = rw.handleBsonM(typedN)
		}

		if fmt.Sprintf("%v", typedN.Type) == "&{bson D}" {
			newNode, _ = rw.handleBsonD(typedN)
		}

		if fmt.Sprintf("%v", typedN.Type) == "&{bson DocElem}" {
			newNode, _ = rw.handleBsonDocElem(typedN)
		}

	default:

	}

	return newNode, rw
}

func rewriteFile(filename string) {
	// parse file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// do the rewrite
	rw := &bsonRewriter{fs: fset}
	newNode := rewriter.Rewrite(rw, node)

	// add import
	// if the imports arent perfect, replace.sh call's goimports
	_ = astutil.AddImport(fset, node, "github.com/10gen/sqlproxy/internal/util/bsonutil")

	// print it back out
	f, err := os.Create(filename)

	if err := format.Node(f, fset, newNode); err != nil {
		log.Fatal(err)
		f.Close()
	}

	f.Close()
}

func main() {
	rewriteFile(os.Args[1])
}
