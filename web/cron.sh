#!/bin/ash

# API endpoint URL
API_URL="http://localhost:3000/api/v1"

# Folder containing files to upload
UPLOAD_FOLDER=$UPLOAD_FOLDER

# Credentials for login
USERNAME=$UPLOAD_USER
PASSWORD=$UPLOAD_PASSWORD

login() {
    local username="$1"
    local password="$2"

    response=$(curl -c cookie.txt --location --request POST "$API_URL/auth/login" --header 'Content-Type: application/json' --data-raw "{\"username\": \"$username\", \"password\": \"$password\"}")

    # Check if login was successful (look for "200 OK" in response headers)
    if [ $? -eq 0 ] && [ "$(echo "$response" | grep -c "token")" -eq 1 ]; then
        echo "Login successful. Cookie obtained."
    else
        echo "Login failed. Unable to obtain cookie."
        exit 1
    fi
}

# Function to upload file and delete if successful
upload_and_delete() {
    local file="$1"

    ls $file
    
    # API call to upload file
    response=$(curl -b cookie.txt --location --request PUT "$API_URL/trail/upload" --header 'Content-Type: application/gpx+xml' --data-binary "@$file")
    
    # Check if API call was successful (status code 200)
    if [ $? -eq 0 ] && [ "$(echo "$response" | grep -c "author")" -eq 1 ]; then
        echo "File $file uploaded successfully."
        # Delete the file
        rm "$file"
        echo "File $file deleted."
    else
        echo $response
        echo "Failed to upload file $file."
    fi
}

# Login to obtain cookie
login "$USERNAME" "$PASSWORD"

# Iterate over each file in the folder
for file in "$UPLOAD_FOLDER"/*; do
    # Check if file exists and is a regular file
    if [ -f "$file" ]; then
        upload_and_delete "$file"
    fi
done
