# Terminal cluster autoscaler

This simple program uses the Terminal.com API to autoscale a Terminal cluster.  

# compilation

```
go build .
```

This will create a binary called `autoscaler`

# usage

For information on how to use it, run
```
./autoscaler --help
```

Example output (this may be somewhat outdated):
```
Usage of ./autoscaler:
  -accesstoken string
        Terminal Cluster user token
  -apiurl string
        Terminal Cluster API Path eg foo.com/api/v0.2/
  -frequency int
        Amount of time to wait between loop of polling cluster state, in seconds (default 5)
  -nodestorage string
        Storage type for nodes to spin up (ebs or ephemeral) (default "ebs")
  -nodetype string
        Name for node type to spin up, e.g. m4.large (default "m4.large")
  -nodetyperam int
        Amount of ram available on an empty node, e.g. 7834 (default 7834)
  -policy string
        The type of autoscaler you want to run general / gpu / other (default "general")
  -tts int
        Minimum time after your cluster grows, in seconds, before shrinking (e.g. if on AWS might as well be 1 hour) (default 3600)
  -usertoken string
        Terminal Cluster user token
```

The user and access token must be that of an admin user on the given cluster

Example usage:
```
./autoscaler -usertoken=$UT -accesstoken $AT -apiurl https://www.terminal.com/api/v0.2 -frequency 5 -nodetype g2.2xlarge -nodetyperam 15000 -tts 3600 -policy=gpu -nodestorage ebs
```
