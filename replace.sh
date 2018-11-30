#pass in name of directory you want to run on
if [ "$1" == "" ]; then
	echo 'Usage: ./replace.sh path/to/directory'
	exit 1
fi 

for f in $(find $1 -name '*.go' -type f});
do

	 #dont want to overwrite function definition
	if [[ $f == *"bsonutil"* ]]; then
		echo Will not run on $f, would overwrite bsonutil.NewX functions
		continue
	fi

	 #dont want to create import cycle
	if [[ $f == *"json"* ]]; then
		echo Will not run on $f, would overwrite calls to mgo bson package
		continue
	fi

	echo $f

	#temporary filename so that go run will not
	#take argument to main.go as another file to run
	fn="${f}_txt"
	tempfile="${f}_tmp"

	mv $f $fn

	go run main.go $fn

	#replace theremin string with newlines
	sed 's/theremin,/\'$'\n''/g' $fn > $tempfile
	cat $tempfile > $fn

	sed 's/theremin/\'$'\n''/g' $fn > $tempfile
	cat $tempfile > $fn

	#rm temp file used for sed
	rm $tempfile

	#back to proper go file
	mv $fn $f

	#clean up file
	goimports -w $f

done

