import json
import requests
import subprocess
import sys
import time
import os
import random
import tempfile
import unittest


class Tests(unittest.TestCase):
    middle_port = 33333

    def setUp(self):
        self.processes = []
        self.cookie = '%x' % random.getrandbits(64)
        self.tmp_dir = tempfile.mkdtemp()
        with open(os.path.join(self.tmp_dir, 'test.txt'), 'w') as f:
            f.write(self.cookie)

    def tearDown(self):
        self.stop_processes()

    def stop_processes(self):
        for cmd, p in self.processes:
            print("terminating %s" % ' '.join(cmd))
            p.terminate()
        for cmd, p in self.processes:
            p.wait()

    def configure(self):
        response = requests.post("http://127.0.0.1:4080/links/", data=json.dumps({
            "port": self.middle_port,
        	"from": {
        		"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDRuinxi4hANygNImiWn6Jjhn5Wyo1tFzmz+x51wvYUNDIHUIdFeX/51yN27+kMv1yUcLvLcbUio925OVan1kFD4VzCfTJ+TqTS4cT8ZnwbrJFZeewFct1aUZeHBB9ttC1WMsXIAA9ZFyFskyN850axiKyvY8Jy4oDedb08OeWRTi+jPjEolD5e33H4JJygujwJxjpdOlbYN+Ah56CcILJXE4O+m5bxy5Krt/hR84+uqOk2aI+8pPVMQxbABPJjaNJZblK9RHGUGuOVAhhA1dW+0rKWoH2bOt6ODW7vggDG0d0G4VwkPvAEWZpkyDroIkk8tHK/jqf9qDi9UsMibVOd",
        		"user": "alice",
        	},
        	"to": {
        		"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCvBNa7e6dJGehmd8KZzgxfrmU/Cyayfd08NpWosT6Je8QNAct+xoU54cT1zYkKnxjME27BG3uF1XGNMW+jZasrh3QJAb8drX2qo65rxhlC5vA7JTQklHkCDiQyOIPtfLGIQCvQQJS3/yjQA59SbFZG4wKS8av8MCS7bW5VP75of9u1T8B8CZAUt3lA+TD6EtYWQFkKJszSOjHbrSLV5PF0QBC+X9kYIXI98ycgOXcXzInssNM7847AtobKNwRqfF83iGkq1C7lMj7dFSpXpUmnvmW41O2cCA/caz1eV1gL/B6JjNBC2FnZC+QtxkMJpi9cPgbqjvLzGEFiQiUNdSf1",
        		"user": "bob",
        	}
        }), headers={
            'Content-Type': 'application/json'
        })
        self.assertEqual(response.status_code, 200)

    def spawn(self, command, **kwargs):
        print("spawn %r" % (command,))
        p = subprocess.Popen(command, **kwargs)
        self.processes.append((command, p))
        return p

    def test_successful_tunnel(self):
        try:
            basedir = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
            self.spawn([os.path.join(basedir, 'sshub'), '--key-file=tests/test_rsa'], cwd=basedir)
            time.sleep(1)
            self.configure()
            subprocess.call('curl http://127.0.0.1:4080/links/ | jq .', shell=True)

            self.spawn(['python', '-m', 'http.server', '8080'], cwd=self.tmp_dir)
            self.spawn(['ssh', '-NL', '9080:localhost:%s' % self.middle_port, '-p', '4022', '-i', 'alice', 'alice@localhost', '-o', 'StrictHostKeyChecking=no'])
            self.spawn(['ssh', '-NR', '%s:localhost:8080' % self.middle_port, '-p', '4022', '-i', 'bob', 'bob@localhost', '-o', 'StrictHostKeyChecking=no'])
            time.sleep(1)
            subprocess.call('curl http://127.0.0.1:4080/tunnels/ | jq .', shell=True)

            response = requests.get('http://127.0.0.1:9080/test.txt')
            result = response.content.decode('ascii')
            subprocess.call('curl http://127.0.0.1:4080/tunnels/ | jq .', shell=True)
        finally:
            self.stop_processes()
        self.assertEqual(result, self.cookie)
