To run this utility, run ./replace.sh <directory>. It will rewrite on that directory recursively.

One issue with this script is that it will 
move comments located before rewritten areas. 

This regex is to find comments within a diff. Comments that you have moved with the script can be found like this.
(\+|-)(\t*)// 

Example: 
-      // This comment was removed 
+      // This comment was added


This will match comments moved to before a close paren:
// (.*)\n\n*\t*\) 

Example: 
bsonutil.NewDocElem(
// This comment got moved
"a", 1
)


This will match comments moved within a bsonutil call: 
// (.*)\n\n*\t*N

Example: 
bsonutil.
// This comment got moved

NewDocElem(
"a", 1
)


