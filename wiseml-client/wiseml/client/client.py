import requests
import time
class Client:

    def __init__(self, url=None):
        self.url = "http://localhost:9021"

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
            time.sleep(0.1)
            logs_resp = requests.post(self.url + "/api/v1/job/logs", json={"JobID": job_id, })
            logs_json = logs_resp.json()
            JobStatus = logs_json["JobStatus"]
            # print("JobStatus: " + JobStatus)
            print(logs_json["Logs"], end="")
