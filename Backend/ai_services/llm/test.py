import os
from azure.ai.inference import ChatCompletionsClient
from azure.ai.inference.models import SystemMessage, UserMessage
from azure.core.credentials import AzureKeyCredential

endpoint = "https://models.github.ai/inference"
model = "openai/gpt-4.1-mini"
token = "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"

client = ChatCompletionsClient(
    endpoint=endpoint,
    credential=AzureKeyCredential(token),
)

response = client.complete(
    messages=[
        SystemMessage(""),
        UserMessage("What is the capital of France?"),
    ],
    temperature=1,
    top_p=1,
    model=model
)

print(response.choices[0].message.content)
