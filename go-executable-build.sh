#!/usr/bin/env bash

package_name=grpcox
version=1.0.0

platforms=("windows/amd64" "windows/386" "darwin/amd64" "linux/amd64")
mkdir -p dist/log
rm -fr dist/index && cp -r index dist/index
rm -fr dist/LICENSE && cp -r LICENSE dist/LICENSE

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$package_name'-'$version'-'$GOOS'-'$GOARCH
    ext=""
    if [ $GOOS = "windows" ]; then
        ext='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o dist/$package_name$ext
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    (cd dist; rm -f $output_name.zip; zip -r $output_name.zip $package_name$ext log index LICENSE)
    rm -f dist/$package_name$ext
done

rm -fr dist/index
rm -fr dist/LICENSE
