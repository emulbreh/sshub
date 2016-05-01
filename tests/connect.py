import json
import requests

response = requests.post("http://127.0.0.1:4080/tunnels/", data=json.dumps({
	"from": {
		"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDRuinxi4hANygNImiWn6Jjhn5Wyo1tFzmz+x51wvYUNDIHUIdFeX/51yN27+kMv1yUcLvLcbUio925OVan1kFD4VzCfTJ+TqTS4cT8ZnwbrJFZeewFct1aUZeHBB9ttC1WMsXIAA9ZFyFskyN850axiKyvY8Jy4oDedb08OeWRTi+jPjEolD5e33H4JJygujwJxjpdOlbYN+Ah56CcILJXE4O+m5bxy5Krt/hR84+uqOk2aI+8pPVMQxbABPJjaNJZblK9RHGUGuOVAhhA1dW+0rKWoH2bOt6ODW7vggDG0d0G4VwkPvAEWZpkyDroIkk8tHK/jqf9qDi9UsMibVOd",
		"user": "alice",
	},
	"to": {
		"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCvBNa7e6dJGehmd8KZzgxfrmU/Cyayfd08NpWosT6Je8QNAct+xoU54cT1zYkKnxjME27BG3uF1XGNMW+jZasrh3QJAb8drX2qo65rxhlC5vA7JTQklHkCDiQyOIPtfLGIQCvQQJS3/yjQA59SbFZG4wKS8av8MCS7bW5VP75of9u1T8B8CZAUt3lA+TD6EtYWQFkKJszSOjHbrSLV5PF0QBC+X9kYIXI98ycgOXcXzInssNM7847AtobKNwRqfF83iGkq1C7lMj7dFSpXpUmnvmW41O2cCA/caz1eV1gL/B6JjNBC2FnZC+QtxkMJpi9cPgbqjvLzGEFiQiUNdSf1",
		"user": "bob",
        "port": 1234,
	}
}), headers={
    'Content-Type': 'application/json'
})
if response.status_code != 200:
    print(response.status_code)
    print(response.content)

