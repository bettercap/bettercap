#!/bin/bash
# nothing to see here, just a utility i use to create new releases ^_^

CURRENT_VERSION=$(cat core/banner.go | grep Version | cut -d '"' -f 2)
TO_UPDATE=(
    core/banner.go
)

echo -n "Current version is $CURRENT_VERSION, select new version: "
read NEW_VERSION
echo "Creating version $NEW_VERSION ...\n"

for file in "${TO_UPDATE[@]}"
do
    echo "Patching $file ..."
    sed -i "s/$CURRENT_VERSION/$NEW_VERSION/g" $file
    git add $file
done

git commit -m "Releasing v$NEW_VERSION"
git push

git tag -a v$NEW_VERSION -m "Release v$NEW_VERSION"
git push origin v$NEW_VERSION

echo
echo "Released on github, building docker image ..."

sudo docker build -t bettercap/bettercap:v$NEW_VERSION .
sudo docker tag bettercap/bettercap:v$NEW_VERSION bettercap/bettercap:latest

echo "Pushing to dockerhub ..."

sudo docker push bettercap/bettercap:v$NEW_VERSION
sudo docker push bettercap/bettercap:latest

echo
echo "All done, v$NEW_VERSION released ^_^"
