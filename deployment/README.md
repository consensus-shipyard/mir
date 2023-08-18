# Deployment

This directory contains tools to deploy Mir on remote machines.
We use [Ansible](https://www.ansible.com/) for configuration and deployment
(and thus ansible must be installed on the local machine to use these tools).
The user only has to provide a list of IP addresses and login credentials to the remote machines,
everything else happens automatically.

## Local setup

1. Save a list of IP addresses of the machines to deploy Mir on in a file, one per line.
   The [`hosts`](hosts) file serves as an example, the contained addresses are meaningless.
2. Configure access to the remote machines by adapting [`group_vars/all.yaml`](group_vars/all.yaml) (see comments in the file).
   Again, the default contents of the file serve as an example, the values are meaningless.

## Remote setup

First, make sure that all remote machines allow passwordless root (sudo) access for the configured user account.
Then, to install Mir and its dependencies on the remote machines, run (from this directory)
```bash
ansible-playbook -i hosts setup.yaml
```
The above command assumes that the remote nodes' IP addresses are stored in `hosts`.
Any file can, however, be used for this purpose.
Variables specified in `group_vars/all.yaml` can be overridden by specifying them as `--extra-vars` arguments.
E.g., to check out a different branch of Mir, say, `bugfix`, run
```bash
ansible-playbook -i hosts setup.yaml --extra-vars "git_version=bugfix"
```
This `setup.yaml` playbook will also compile the sample chat demo application
and the code for preliminary performance evaluation.

## Deploying applications

Currently, the chat demo application and the simple benchmarking app are supported out-of-the-box.
To start them on the remote machines, run, respectively
```bash
ansible-playbook -i hosts start-chat-demo.yaml
```
or
```bash
ansible-playbook -i hosts run-benchmark.yaml
```

Each of the commands above will start the respective application on all machines specified in `hosts`.
Moreover, the `start-benchmark.yaml` playbook will create a `membership` file in the current directory,
containing membership configuration in Mir's format, which can be used to start a local bench client.

The application is run on the remote machines within a `tmux` session.
To see the console output of the application, log in to the remote machine and run
```bash
tmux a
```
To attach to the tmux session.

## Stopping nodes

To stop the application on the remote machines, run
```bash
ansible-playbook -i hosts stop-nodes.yaml
```
This will kill the tmux server on each remote machine.

## Running a custom script

To run a custom script on the remote machines, modify the [`scripts/custom.sh`](scripts/custom.sh) file and run
```bash
ansible-playbook -i hosts custom-script.yaml
```
This will execute [`scripts/custom.sh`](scripts/custom.sh) on each remote machine.
