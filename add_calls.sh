#!/bin/zsh

usage() {
	echo
	echo "./squawker_script"
	echo "\t-r, --randomize\tRandomizes song order"
}

squawker_endpoint="http://localhost:15567/add?video="
input=$(cat video_ids)

while [ "$1" != "" ]; do
    PARAM=`echo $1 | awk -F= '{print $1}'`
    VALUE=`echo $1 | awk -F= '{print $2}'`
    case $PARAM in
        -h | --help)
            usage
            exit
            ;;
        -r | --randomize)
            input=$(cat video_ids | shuf)
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            usage
            exit 1
            ;;
    esac
    shift
done

ids=("${(f@)$(echo $input)}")
for id in "${ids[@]}"
do
	echo "$id"
	curl "$squawker_endpoint$id" | jq; sleep 0.25
done
