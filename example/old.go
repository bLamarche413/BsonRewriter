package main

import (
	"github.com/10gen/sqlproxy/internal/util/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

func main() {

	docElem := bson.DocElem{Name: "a", Value: 1}

	bsonM := bson.M{"apple": 1}

	bsonM2 := bson.M{
		"apple":    1,
		"banana":   2,
		"cucumber": 3,
	}

	bsonMNested := bson.M{
		"apple": bson.M{
			"banana": bson.M{
				"cucumber": bson.M{
					"daikon": 2}}}}

	bsonD := bson.D{{Name: "bar", Value: 2}}

	bsonD2 := bson.D{
		{Name: "c_DOT_a", Value: "$c.a"},
		{Name: "c_DOT_d", Value: "$c.d"}}

	bsonMArray := []bson.M{
		{
			"first":  bson.M{"hello": "world"},
			"second": "What",
		},
		{
			"town": "Lavalette",
		},
	}

	bsonDArray := []bson.D{
		{{Name: "a", Value: bson.D{{Name: "b", Value: 7}}},
			{Name: "b", Value: 7}, {Name: "_id", Value: 5}},
		{{Name: "a", Value: 16}, {Name: "b", Value: 17}, {Name: "_id", Value: 15}},
	}

	interfaceInBson := bson.DocElem{Name: "a", Value: []interface{}{"a", "b", 1, 2}}

	interfaceininterface := bson.DocElem{Name: "a", Value: []interface{}{"a", "b", 1, 2, []interface{}{"inside", "array"}}}

	callexprM := bson.M{
		bsonutil.NewDocElem("a", 1),
	}

	callexprD := bson.D{
		bsonutil.NewDocElem("a", 1),
	}

	callexprDArray := []bson.D{

		bsonutil.NewD(bsonutil.NewDocElem("a", 1)),
	}

	callExprMArray := []bson.M{
		bsonutil.NewM(bsonutil.NewDocElem("a", 1)),
	}

}
