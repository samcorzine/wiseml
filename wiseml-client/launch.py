from wiseml.tensorflow import Tensorflow

# Point to the directory containing the source code for the training job
tf_estimator = Tensorflow('train.py', source_dir='src')

# Starts a wiseml training job, currently does not use the input_data
tf_estimator.fit('s3://my_bucket/my_training_data/')