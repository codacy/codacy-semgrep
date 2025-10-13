# Before (non-Guardrails)
from openai import OpenAI
client = OpenAI()
resp = client.responses.create(model="gpt-4.1", input="Hello")

openai.ChatCompletion.create(model="gpt-4.1", input="Hello")

# After (Guardrails)
from guardrails import GuardrailsOpenAI
client = GuardrailsOpenAI(config="guardrails_config.json")
resp = client.responses.create(model="gpt-4.1", input="Hello")
# Then check resp.guardrail_results, etc.