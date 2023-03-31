import requests
import time
import yaml
import os


def load_config(path) -> dict:
    with open(path) as f:
        config = yaml.load(f, Loader=yaml.FullLoader)
    return config

class Client:

    def __init__(self, **kwargs):
        self.url = kwargs.get("url", None)
        if not self.url:
            conf = load_config(os.path.expanduser("~/.config/wiseml.yaml"))
            self.url = conf["server-url"]




    def launch_training_job(self, source_path: str):
        endpoint = "/api/v1/job/launch"
        url = self.url + endpoint
        r = requests.post(url, files={"file": open(source_path, "rb")})
        print(r.status_code)
        body = r.json()
        job_id = body["JobID"]
        print("JobID: " + job_id)
        JobStatus = "UNKNOWN"
        while JobStatus != "Completed":
            time.sleep(1)
            logs_resp = requests.post(self.url + "/api/v1/job/logs", json={"JobID": job_id, })
            logs_json = logs_resp.json()
            JobStatus = logs_json["JobStatus"]
            # print("JobStatus: " + JobStatus)
            print(logs_json["Logs"], end="")
