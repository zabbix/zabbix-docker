import sys
import requests
import json
import markdown
import os

repository_description = ''

if ("DESCRIPTION_FILE" not in os.environ):
    print("Description file environment variable is not specified")
    sys.exit(1)

if (os.path.isfile(os.environ["DESCRIPTION_FILE"] + '.md')):
    repository_description=markdown.markdownFromFile(input=os.environ["DESCRIPTION_FILE"] + '.md')
elif (os.path.isfile(os.environ["DESCRIPTION_FILE"] + '.html')):
    file = open(os.environ["DESCRIPTION_FILE"] + '.html', mode='r')
    repository_description = file.read()
    file.close()

if (len(repository_description)) == 0:
    print("No description")
    sys.exit(1)

data = dict()
data['container'] = dict()
data['container']['repository_description'] = repository_description[:32768]

headers = {'accept' : 'application/json', 'X-API-KEY' : os.environ["PYXIS_API_TOKEN"], 'Content-Type' : 'application/json'}
result = requests.patch(os.environ["API_URL"] + os.environ["PROJECT_ID"], headers = headers, data = json.dumps(data))
print(result)
print(json.loads(r.content)['last_update_date'])