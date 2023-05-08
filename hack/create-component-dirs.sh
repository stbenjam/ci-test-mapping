#!/bin/bash
# This creates the component directory structure from a dump from JIRA
# using this command. You'll need to set TOKEN to your personal access
# token from JIRA.
#
# curl -H "Accept: application/json" \
#     -X GET \
#     -H "Authorization: Bearer $(echo -n $TOKEN)" \
#     -H "Content-Type: application/json" \
#     "https://issues.redhat.com/rest/api/2/issue/createmeta/OCPBUGS/issuetypes/1" \
#     | jq -r '[.values[] | select(.fieldId == "components") | .allowedValues[].name]' \
#     | jq 'reduce .[] as $item ({}; if ($item | contains("/")) then ($item | split(" / ")) as $key_val | (.[$key_val[0]] += [$key_val[1]]) else .[$item] = [] end)'

set -o errexit
set -o pipefail

# Ensure jq is installed
if ! command -v jq >/dev/null 2>&1; then
  echo "The 'jq' command is required for this script. Please install it and try again."
  exit 1
fi

# Check if a JSON file is provided
if [[ -z "$1" ]]; then
  echo "Usage: $0 <JSON_FILE> [PARENT_PATH]"
  exit 1
fi

JSON_FILE="$1"
PARENT_PATH="${2:-./}"

# Ensure PARENT_PATH has a trailing '/'
[[ "${PARENT_PATH: -1}" != '/' ]] && PARENT_PATH+="/"

create_files() {
  local path="$1"
  local component="$2"
  local package_name="$3"

  local example
  example=$(dirname "$0")/../pkg/components/example

  # Create OWNERS file
  echo "component: ${component}" > "${path}/OWNERS"

  # Copy example files
  cp "$example"/*.go "$path"

  camelCase=$(echo "$component" | sed 's/([^)]*)//g' | tr ' /-' '\n' | awk '{printf "%s%s", toupper(substr($0,1,1)), substr($0,2)}' | tr '\n' ' ' | sed 's/ $//')

  registry_file="$example/../../registry/registry.go"
  sed -i -e "s#\(import (\)#\1\n\t\"github.com/openshift-eng/ci-test-mapping/${path:2}\"#g" "$registry_file"
  sed -i -e "s#\(// New components go here\)#\tregistry.Register(\"$component\", \&${package_name}.${camelCase}Component)\n\1#g" "$registry_file"

  # Loop over all .go files in the directory
  for file in "$path"/*.go; do
    sed -i -e "s/ExampleComponent/${camelCase}Component/g" "$file"
    sed -i -e "s/package example/package ${package_name}/g" "$file"
    sed -i -e "s#\"Example\"#\"$component\"#g" "$file"
  done
}

# Function to create directory structure recursively
create_structure() {
  local parent_path="$1"
  local json_obj="$2"
  local prefix_path="$3"

  while IFS= read -r dir; do
    [[ "$dir" =~ "Documentation" ]] && continue

    package_name=$(echo "$dir" | sed 's/([^)]*)//g' | tr -d '[:punct:][:space:]' | tr '[:upper:]' '[:lower:]')

    # Create the directory if it doesn't exist
    current_path="${parent_path}${package_name}"
    mkdir -p "$current_path"

    # Populate example files
    create_files "${current_path}" "${prefix_path}${dir}" "${package_name}"

    # Get the subdirectories for the current directory
    subdirs=$(jq ".[\"$dir\"]" <<< "$json_obj")

    # Check if subdirs is an object or an array
    if [[ $(jq 'if type=="object" then true else false end' <<< "$subdirs") == "true" ]]; then
      # Recursively create the directory structure for subdirectories
      create_structure "${current_path}/" "$subdirs" "${prefix_path}${dir} / "
    elif [[ $(jq 'if type=="array" then true else false end' <<< "$subdirs") == "true" ]]; then
      # If subdirs is an array, create the directories for plain strings
      while IFS= read -r subdir; do
        package_subdir=$(echo "$subdir" | sed 's/([^)]*)//g' | tr -d '[:punct:][:space:]' | tr '[:upper:]' '[:lower:]')
        mkdir -p "${current_path}/${package_subdir}"
        create_files "${current_path}/${package_subdir}" "${prefix_path}${dir} / ${subdir}" "${package_subdir}"
      done < <(jq -r '.[]' <<< "$subdirs")
    fi
  done < <(jq -r 'keys[]' <<< "$json_obj")
}

# Read JSON content from the file
json_content=$(cat "$JSON_FILE")

# Create directory structure from JSON
create_structure "$PARENT_PATH" "$json_content" ""

echo "Directory structure created successfully."

