from security_utils import decode_access_token

token = "4uo3TBPh48tUm-9TxmqtRYuSVYqxaGHj3RQxj8sRLvU"
payload = decode_access_token(token)

print("Decoded Payload:", payload)
