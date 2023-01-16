#! /bin/bash
cd ..

PROJ_NAME=levelqueue
GITHUB_PATH='github.com/pengdacn/levelqueue'
COMPANY_PATH='gitlab.test.com/common/levelqueue'

mkdir -p company/$PROJ_NAME

for f in company/$PROJ_NAME/*; do
  rm -rf "$f"
done

for f in ./$PROJ_NAME/*; do
  cp -r "$f" company/$PROJ_NAME/
done

cd company/$PROJ_NAME

for f in $(find -name '*.go'); do
  sed -i "s,${GITHUB_PATH},${COMPANY_PATH},g" $f
done

top_dir=$(pwd)
for f in $(find -name 'go.mod'); do
  local_proj=$(dirname $f)

  cd $local_proj
  go mod tidy

  cd $top_dir
done
