from wiseml.tensorflow import Tensorflow

# Configure an MXNet Estimator (no training happens yet)
tf_estimator = Tensorflow('train.py', source_dir='src')

# Starts a SageMaker training job and waits until completion.
tf_estimator.fit('s3://my_bucket/my_training_data/')