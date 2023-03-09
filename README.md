### Wise ML

WiseML is a system that allows you to turn your gaming pc into a Sagemaker-compatible machine learning server. It currently supports almost none of the features of Sagemaker, but it's a start.

It has two components right now, a server and a client. The server is a Golang backend that is meant to run on your gaming pc. It's job is to listen for requests from the client and then run the requested code on your gaming pc. The client is a Python library that is meant to be used on your laptop. It's job is to send requests to the server and then return the results.

#### Installation

To install the server, you need to have Golang and make installed. From there you should be able to run

```
make server
```


Once that is up and running you should be able to run these commands to launch a test job.

```
cd wiseml-client
python launch.py
```

There's a good chance you will run into problems, if so please let me know with a github issue or otherwise!