from os.path import dirname

import requests
import yaml


def fixture(target):
    r = requests.get("%s/_classes" % target)
    print(r.json())
    for f in yaml.load(open('%s/fixture.yml' % dirname(__file__), 'r')):
        r = requests.post("%s/project" % target, json=f)
        assert r.status_code == 201, r.status_code
        print(r.json())


if __name__ == "__main__":
    fixture("http://127.0.0.1:5000")
