[![CircleCI](https://circleci.com/gh/disaster37/check_elasticsearch/tree/7.x.svg?style=svg)](https://circleci.com/gh/disaster37/check_elasticsearch/tree/7.x)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/check_elasticsearch)](https://goreportcard.com/report/github.com/disaster37/check_elasticsearch)
[![GoDoc](https://godoc.org/github.com/disaster37/check_elasticsearch?status.svg)](http://godoc.org/github.com/disaster37/check_elasticsearch)
[![codecov](https://codecov.io/gh/disaster37/check_elasticsearch/branch/7.x/graph/badge.svg)](https://codecov.io/gh/disaster37/check_elasticsearch/branch/7.x)


# check_elasticsearch
Nagios plugin to check the healtch of Elasticsearch cluster

## Contribute

You PR are always welcome. Please use the righ branch to do PR:
 - 7.x for Elasticsearch 7.x
Don't forget to add test if you add some functionalities.

To build, you can use the following command line:
```sh
go build
```

To lauch golang test, you can use the folowing command line:
```sh
make test
```

## CLI

### Global options

The following parameters are available for all commands line :
- **--url**: The Elasticsearch URL. For exemple https://elasticsearch.company.com. Alternatively you can use environment variable `ELASTICSEARCH_URL`.
- **--user**: The login to connect on Elasticsearch. Alternatively you can use environment variable `ELASTICSEARCH_USER`.
- **--password**: The password to connect on Elasticsearch. Alternatively you can use environment variable `ELASTICSEARCH_PASSWORD`.
- **--self-signed-certificate**: Disable the check of server SSL certificate
- **--debug**: Enable the debug mode
- **--help**: Display help for the current command


You can set also this parameters on yaml file (one or all) and use the parameters `--config` with the path of your Yaml file.
```yaml
---
url: https://elasticsearch.company.com
user: elastic
password: changeme
```

### Check if indice are locked by storage pressure

Command `check-indice-locked` permit to check if indice provided is not locked by storage pressure.
If you should to check all indice, you can put `_all` as indice name.

You need to set the following parameters:
- **--indice**: The indice name to check

It return the following perfdata:
- **nbIndices**: the number of indices returned
- **nbIndicesLocked**: the number of indices locked


Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-indice-locked --indice _all
```

Response:
```bash
OK - No indice locked (6/6)|nbIndices=6;;;; nbIndicesLocked=0;;;;
```

### Check that ILM service is running

Command `check-ilm-status` permit to check if ILM service is running


Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-ilm-status
```
Response:
```bash
OK - ILM is running
```


### Check ILM errors on indice

Command `check-ilm-indice` permit to check if ILM policy failed on given indice.
If you should to check all indice, you can put `_all` as indice name.

You need to set the following parameters:
- **--indice**: The indice name
- **--exclude**: (optional) The indice name you should to exclude

It return the following perfdata:
- **nbIndicesFailed**: the number of indices with ILM error

Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-ilm-indice --indice _all
```

Response:
```bash
OK - No error found on indice _all|NbIndiceFailed=0;;;; 
```

### Check that SLM service is running 

Command `check-slm-status` permit to check if SLM service is running


Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-slm-status
```
Response:
```bash
OK - SLM service is running
```

### Check if there are snapshot errors

Command `check-repository-snapshot` permit to check if there are snapshot error on given repository.

You need to set the following parameters:
- **--repository**: The repository name where you should to check snapshots

It return the following perfdata:
- **nbSnapshot**: the number of snapshot
- **nbSnapshotFailed**: the number of failed snapshot

Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-repository-snapshot --repository snapshot
```

Response:
```bash
OK - No snapshot on repository snapshot|NbSnapshot=0;;;; NbSnapshotFailed=0;;;;
```

### Check Transform errors

Command `check-transform` permit to check if tranform failed.
If you should to check all tranform, you can put `_all` as transform name.

You need to set the following parameters:
- **--name**: The transform name
- **--exclude**: (optional) The transform name you should to exclude

It return the following perfdata:
- **nbTransformFailed**: the number of transform failed
- **nbTransformStarted**: the number of transform started
- **nbTransformStopped**: the number of transform stopped

Sample of command:
```bash
./check_elasticsearch --url http://localhost:9200 --user elastic --password changeme check-transform --name _all
```

Response:
```bash
OK - No error found on indice _all|NbIndiceFailed=0;;;; 
```
