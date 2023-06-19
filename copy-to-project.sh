#!/bin/bash

# Get the directory of the script
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# Default values
destination_dir=""
new_module_name=""

# Function to print usage information
usage() {
    echo -e "Usage: ./copy_files.sh \n\t-t <destination_directory> \n\t-m <new_module_name> \n\t-n <new_project_name> optional"
}

# Parse command line arguments
while getopts ":t:m:n:" opt; do
    case $opt in
    t)
        destination_dir="${OPTARG%/}" # Remove trailing slash, if any
        ;;
    m)
        new_module_name="$OPTARG"
        ;;
    n)
        new_project_name="$OPTARG"
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
if [[ -z $destination_dir || -z $new_module_name ]]; then
    echo "Missing required argument(s)."
    usage
    exit 1
fi

if [ ! -d "$destination_dir" ]; then
    mkdir -p $destination_dir
fi


if [[ -z $new_project_name ]]; then
    new_project_name=$(echo "$new_module_name" | awk -F '/' '{print $NF}')
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

    shopt -s dotglob # Include hidden files and folders

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
            replace_in_file "$new_module_name" "$dest_file"
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
        replace_in_file "$new_module_name" "$dest_gitignore"
    fi

    shopt -u dotglob # Disable including hidden files and folders
}

# Replace string in a file
replace_in_file() {
    local replace_str="$1"
    local file="$2"

    # Find and replace the string in the file
    sed -i "s#$old_module#$new_module_name#g" "$file"
    sed -i "s#Skeleton#$new_project_name#gI" "$file"
}

# Call the function to start copying files
copy_files "$script_dir" "$destination_dir"

