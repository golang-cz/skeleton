#!/bin/bash

# Get the directory of the script
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# Default values
destination_dir=""
replace_string=""

# Function to print usage information
usage() {
    echo "Usage: ./copy_files.sh -t <destination_directory> -m <replace_string> -n <name-of-new-project>"
}

# Parse command line arguments
while getopts ":t:m:n:" opt; do
    case $opt in
    t)
        destination_dir="${OPTARG%/}" # Remove trailing slash, if any
        ;;
    m)
        replace_string="$OPTARG"
        ;;
    n)
        new_name="$OPTARG"
        ;;
    \?)
        echo "Invalid option: -$OPTARG" >&2
        usage
        exit 1
        ;;
    :)
        echo "Option -$OPTARG requires an argument." >&2
        usage
        exit 1
        ;;
    esac
done

# Check if destination directory is provided
if [[ -z $destination_dir || -z $replace_string || -z $new_name ]]; then
    echo "Missing required argument(s)."
    usage
    exit 1
fi

# template module name
old_module=$(cat go.mod | grep "module " | awk -F'module ' '{print $2}')

# Function to check if a file or folder is ignored
is_ignored() {
    local file=$1
    # Check if the file matches any pattern in .gitignore
    if git -C "$script_dir" check-ignore -q "$file"; then
        return 0
    else
        return 1
    fi
}

# Copy files recursively, excluding ignored files
copy_files() {
    local source=$1
    local destination=$2

    for file in "$source"/*; do
        local file_name=$(basename "$file")

        # Skip .git folder and the script itself
        if [[ "$file_name" == ".git" || "$file_name" == "$(basename "${BASH_SOURCE[0]}")" ]]; then
            continue
        fi

        # Skip ignored files
        if is_ignored "$file"; then
            continue
        fi

        # Copy non-ignored files to the destination directory
        if [[ -f "$file" ]]; then
            local dest_file="$destination/$file_name"
            cp "$file" "$dest_file"
            replace_in_file "$replace_string" "$dest_file"
        elif [[ -d "$file" ]]; then
            local dest_dir="$destination/$file_name"
            mkdir -p "$dest_dir"
            copy_files "$file" "$dest_dir"
        fi
    done

    # Copy .gitignore file if it exists
    local gitignore_file="$source/.gitignore"
    if [[ -f "$gitignore_file" ]]; then
        local dest_gitignore="$destination/.gitignore"
        cp "$gitignore_file" "$dest_gitignore"
        replace_in_file "$replace_string" "$dest_gitignore"
    fi
}

# Replace string in a file
replace_in_file() {
    local replace_str="$1"
    local file="$2"

    # Find and replace the string in the file
    sed -i "s#$old_module#$replace_string#g" "$file"
    sed -i "s#Skeleton#$new_name#gI" "$file"
}

# Call the function to start copying files
copy_files "$script_dir" "$destination_dir"

