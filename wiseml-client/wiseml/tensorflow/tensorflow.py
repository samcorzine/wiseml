import time
from wiseml.utils import *
from wiseml.client import Client
import requests

class Tensorflow:
    def __init__(self, entry_point, source_dir, **kwargs):
        self.entry_point = entry_point
        self.source_dir = source_dir
        self.client = Client(url="http://localhost:9021")

    def fit(self, input_data):
        print("Tensorflow.fit() called with input_data: " + input_data)
        path = tar_from_dir(self.source_dir)
        self.client.launch_training_job(path)
