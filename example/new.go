package main

import (
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
)

func main() {

	docElem := bsonutil.NewDocElem("a", 1)

	bsonM := bsonutil.NewM(bsonutil.NewDocElem("apple", 1))

	bsonM2 := bsonutil.NewM(
		bsonutil.NewDocElem("apple", 1),
		bsonutil.NewDocElem("banana", 2),
		bsonutil.NewDocElem("cucumber", 3),
	)

	bsonMNested := bsonutil.NewM(
		bsonutil.NewDocElem("apple", bsonutil.NewM(
			bsonutil.NewDocElem("banana", bsonutil.NewM(
				bsonutil.NewDocElem("cucumber", bsonutil.NewM(
					bsonutil.NewDocElem("daikon", 2))))))))

	bsonD := bsonutil.NewD(bsonutil.NewDocElem("bar", 2))

	bsonD2 := bsonutil.NewD(
		bsonutil.NewDocElem("c_DOT_a", "$c.a"),
		bsonutil.NewDocElem("c_DOT_d", "$c.d"),
	)

	bsonMArray := bsonutil.NewMArray(
		bsonutil.NewM(
			bsonutil.NewDocElem("first", bsonutil.NewM(bsonutil.NewDocElem("hello", "world"))),
			bsonutil.NewDocElem("second", "What"),
		),
		bsonutil.NewM(
			bsonutil.NewDocElem("town", "Lavalette"),
		),
	)

	bsonDArray := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", bsonutil.NewD(bsonutil.NewDocElem("b", 7))),
			bsonutil.NewDocElem("b", 7),
			bsonutil.NewDocElem("_id", 5),
		),
		bsonutil.NewD(bsonutil.NewDocElem("a", 16), bsonutil.NewDocElem("b", 17), bsonutil.NewDocElem("_id", 15)),
	)

	interfaceInBson := bsonutil.NewDocElem("a", bsonutil.NewArray(
		"a",
		"b",
		1,
		2,
	))

	interfaceininterface := bsonutil.NewDocElem("a", bsonutil.NewArray(
		"a",
		"b",
		1,
		2,
		bsonutil.NewArray(
			"inside",
			"array",
		),
	))

	callexprM := bsonutil.NewM(
		bsonutil.NewDocElem("a", 1),
	)

	callexprD := bsonutil.NewD(
		bsonutil.NewDocElem("a", 1),
	)

	callexprDArray := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("a", 1)),
	)

	callExprMArray := bsonutil.NewMArray(
		bsonutil.NewM(bsonutil.NewDocElem("a", 1)),
	)

}
