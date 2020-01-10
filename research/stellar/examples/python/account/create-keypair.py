from stellar_sdk.keypair import Keypair

kp = Keypair.random()
print("Key: {}".format(kp.secret))
print("Address: {}".format(kp.public_key))
