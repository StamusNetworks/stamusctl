# Check if two arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <folder1> <folder2>"
    exit 1
fi

folder1=$1
folder2=$2

# Check if both arguments are directories
if [ ! -d "$folder1" ]; then
    echo "Error: $folder1 is not a directory"
    exit 1
fi

if [ ! -d "$folder2" ]; then
    echo "Error: $folder2 is not a directory"
    exit 1
fi

# Find all files in both folders and sort them
files1=$(find "$folder1" -type f | sed "s|^$folder1/||" | sort)
files2=$(find "$folder2" -type f | sed "s|^$folder2/||" | sort)

# Compare the list of files
diff <(echo "$files1") <(echo "$files2") > /dev/null
if [ $? -ne 0 ]; then
    echo "The folders have different files."
    echo "Files in $folder1 but not in $folder2:"
    comm -23 <(echo "$files1") <(echo "$files2")
    echo "Files in $folder2 but not in $folder1:"
    comm -13 <(echo "$files1") <(echo "$files2")
    exit 1
fi

# Check the contents of each file
for file in $files1; do
    diff <(tr -d '\n' < "$folder1/$file") <(tr -d '\n' < "$folder2/$file") > /dev/null
    if [ $? -ne 0 ]; then
        echo "Files $folder1/$file and $folder2/$file differ"
        echo "Diff:"
        diff <(tr -d '\n' < "$folder1/$file") <(tr -d '\n' < "$folder2/$file")
        exit 1
    fi
done

echo "The folders are identical."
exit 0