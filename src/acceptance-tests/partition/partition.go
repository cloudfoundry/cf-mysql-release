package partition

import (
	"bytes"
	"fmt"
	"log"

	"code.google.com/p/go.crypto/ssh"
)

const password = "c1oudc0w"

func On(ssh_tunnel string, local_ip string) {
	session := sshSession(ssh_tunnel)
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	remoteCommands := fmt.Sprintf(`echo %[1]s | sudo -S true && \
    sudo iptables -A INPUT  ! -s 127.0.0.1 -p tcp ! --destination-port 22 -j DROP && \
    sudo iptables -A OUTPUT   -s 127.0.0.1 -p tcp ! --source-port 22 -j DROP && \
    sudo iptables -A INPUT  ! -s %[2]s     -p tcp ! --destination-port 22 -j DROP && \
    sudo iptables -A OUTPUT   -s %[2]s     -p tcp ! --source-port 22 -j DROP
    `, password, local_ip)

	println(remoteCommands)

	err := session.Run(remoteCommands)
	if err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}

func Off(ssh_tunnel string) {
	session := sshSession(ssh_tunnel)
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	err := session.Run(fmt.Sprintf("echo %[1]s | sudo -S iptables -F", password))
	if err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}

func sshSession(ip string) *ssh.Session {
	config := &ssh.ClientConfig{
		User: "vcap",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", ip, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}

	return session
}
