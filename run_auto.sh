set -v

cd $( dirname "${BASH_SOURCE[0]}" )
. vars.sh
#rm autoscaler
#go build .

##PUT USER AND AUTH TOKEN HERE IN PLACE OF THESE SAMPLE CREDS##
#USER_TOKEN=6aef1a568dc37b1e166d8c59f95d2f921575d6a20abc58e61a75fa7c6aa10b2d
#ACCESS_TOKEN=5fs6nsfrfnnru5i1gsc5oaio4bhsirsl2dshahu4
#DOMAIN=ccterminalcloud.com
#URL=https://www.$DOMAIN/api/v0.2/
  

# don't shrink for 1 hour after starting
#TIME_TO_SHRINK=3600
#NODE_TYPE=r3.4xlarge # c4.8xlarge, c4.4xlarge
#NODE_TYPE_RAM=15000 # 60000, 30000 # units = megabytes
while true
do
  ./slackpost $(cat slack_token) internal_codecademy "starting autoscaler"
  autoscalerd -usertoken=$USER_TOKEN -accesstoken $ACCESS_TOKEN -apiurl $URL -frequency 20 -nodetype r3.4xlarge -nodetyperam 122792 -tts 60 -policy=general -nodestorage ephemeral 2>&1 | tee  autoscaler.log
done
