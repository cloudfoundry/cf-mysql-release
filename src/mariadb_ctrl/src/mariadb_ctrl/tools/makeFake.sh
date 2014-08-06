#! /bin/bash


# Makes a fake for a go interface
# compensating for the fact that we don't have a fully formed go path
# USAGE:  TODO


if [ -z "$1" ]
  then
    echo "No argument supplied"
fi

CURRENT_DIR=`pwd`
GO_COPY_DIR=/Users/pivotal/go/src/tmp
rm -rf $GO_COPY_DIR
mkdir -p $GO_COPY_DIR
cp -r * $GO_COPY_DIR

FAKE_DIR=$GO_COPY_DIR/$1
cd $FAKE_DIR
echo "pwd: `pwd`"
counterfeiter . $2

cp -r $FAKE_DIR/fakes $CURRENT_DIR/$1

rm -rf $GO_COPY_DIR
