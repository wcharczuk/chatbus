#! /bin/sh

: ${ROOT:?"ROOT is required"}

echo "mode: set" > ./profile.cov

for dir in $(find ${ROOT} -maxdepth 10 -not -path '*/testdata' -not -path '*/vendor/*' -type d);
do
if ls $dir/*.go &> /dev/null; then
	go test -short -covermode=set -coverprofile=$dir/profile.tmp $dir
	if [ -f $dir/profile.tmp ]; then
		cat $dir/profile.tmp | tail -n +2 >> profile.cov
		rm $dir/profile.tmp
	fi
fi
done

go tool cover -html=profile.cov

rm profile.cov