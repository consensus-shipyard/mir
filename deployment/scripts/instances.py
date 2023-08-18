import pathlib
import queue
import threading

import sys
import boto3
import yaml
from botocore.exceptions import ClientError

DEFAULT_REGION_FILE = "regions.yaml"

regions = {}
ec2_clients = {}


def load_regions(file_name):
    global regions
    global ec2_clients

    # Load region data
    with open(file_name, 'r') as yaml_file:
        regions = yaml.safe_load(yaml_file)

    # Initialize EC2 clients.
    for region in regions.keys():
        ec2_clients[region] = boto3.client('ec2', region_name=region)


def run_concurrently(f, args):
    threads = []
    results = queue.Queue()

    def wrapper_func(target, region_name, result_queue, target_args):
        result_queue.put((region_name, target(region, *target_args)))

    # Run function concurrently for each region
    for region in ec2_clients.keys():
        if isinstance(args, dict):
            thread = threading.Thread(target=wrapper_func, args=(f, region, results, args[region]))
        else:
            thread = threading.Thread(target=wrapper_func, args=(f, region, results, args))
        thread.start()
        threads.append(thread)

    # Wait for results.
    for t in threads:
        t.join()

    # Gather results
    res = {}
    while not results.empty():
        result = results.get()
        res[result[0]] = result[1]
    return res


def get_name_tag(inst):
    for tag in inst['Tags']:
        if tag['Key'] == 'Name':
            return tag['Value']
    return ""


def get_public_ip(inst):
    if 'PublicIpAddress' in inst:
        return inst['PublicIpAddress']
    else:
        return ""


def get_instance_ids(region_name, name_tag):

    result = ec2_clients[region_name].describe_instances(Filters=[{'Name': 'tag:Name', 'Values': [name_tag]}])
    instance_ids = []

    for res in result['Reservations']:
        for inst in res['Instances']:
            if inst['State']['Name'] not in ["terminated", "shutting-down"]:
                instance_ids.append(inst['InstanceId'])

    return instance_ids


def get_all_instance_ids(name_tag):
    return run_concurrently(get_instance_ids, [name_tag])


def get_instance_info(region_name, instance_ids):
    if len(instance_ids) == 0:
        return "", 0

    result = ec2_clients[region_name].describe_instances(InstanceIds=instance_ids)
    count = 0
    data = ""
    for res in result['Reservations']:
        for inst in res['Instances']:
            count += 1
            data += f"{region_name:<15}"+" "+inst['InstanceId']+" "+f"{inst['InstanceType']:<15}"+" "+f"{get_name_tag(inst):<20}"+" ("+f"{get_public_ip(inst):<15}"+"): "+inst['State']['Name']+"\n"
    return data, count


def print_instances(instance_ids):
    instance_info = run_concurrently(get_instance_info, {region_name: [ids] for region_name, ids in instance_ids.items()})

    total_count = 0
    for region_name, inst_data in instance_info.items():
        print(inst_data[0], end="")
        total_count += inst_data[1]
    print(f"Total instances: {total_count}")


def print_all_instances(name_tag):
    print_instances(get_all_instance_ids(name_tag))


def confirm_instances(instance_ids, message):
    print_instances(instance_ids)
    confirmation = input(message+" Proceed? (yes/no): ")
    return confirmation.lower() == "yes"


def create_instances(region_name, name_tag, count, instance_type):
    tag_spec = {
        'ResourceType': 'instance',
        'Tags': [{
            'Key': 'Name',
            'Value': name_tag
        }]
    }
    disk_size = 20
    disk_spec = {
        'DeviceName': '/dev/sda1',  # Device name on the instance
        'Ebs': {
            'VolumeSize': disk_size,  # Size in GiB
            'VolumeType': 'gp3',  # EBS volume type (e.g., gp2, io1, standard)
            'DeleteOnTermination': True  # Delete volume on instance termination
        }
    }

    result = ec2_clients[region_name].run_instances(
        MinCount=count,
        MaxCount=count,
        ImageId=regions[region_name]["ami-id"],
        InstanceType=instance_type,
        KeyName=regions[region_name]["key-name"],
        SecurityGroupIds=[regions[region_name]["security-group"]],
        BlockDeviceMappings=[disk_spec],
        TagSpecifications=[tag_spec],
    )

    # Wait until instances are started
    instance_ids = []
    for inst in result['Instances']:
        instance_ids.append(inst['InstanceId'])
    waiter = ec2_clients[region_name].get_waiter('instance_running')
    waiter.wait(InstanceIds=instance_ids)

    return instance_ids


def create_all_instances(name_tag, count, instance_type):
    instance_ids = run_concurrently(create_instances, [name_tag, count, instance_type])
    print_instances(instance_ids)


def terminate_instances(region_name, instance_ids):
    # Do a dryrun first to verify permissions
    try:
        ec2_clients[region_name].terminate_instances(InstanceIds=instance_ids, DryRun=True)
    except ClientError as e:
        if 'DryRunOperation' not in str(e):
            raise

    try:
        ec2_clients[region_name].terminate_instances(InstanceIds=instance_ids, DryRun=False)
    except ClientError as e:
        print(e)

    # Wait until instances are terminated.
    waiter = ec2_clients[region_name].get_waiter('instance_terminated')
    waiter.wait(InstanceIds=instance_ids)


def terminate_all_instances(name_tag):
    instance_ids = get_all_instance_ids(name_tag)

    # Dry run succeeded, confirm decision and run terminate_instances without dryrun
    if not confirm_instances(instance_ids, "These instances will be TERMINATED."):
        print("Aborting.")
        return
    print("Terminating...")

    run_concurrently(terminate_instances, {region_name: [ids] for region_name, ids in instance_ids.items()})
    print_instances(instance_ids)


def start_instances(region_name, instance_ids):

    if len(instance_ids) == 0:
        return

    # Do a dryrun first to verify permissions
    try:
        ec2_clients[region_name].start_instances(InstanceIds=instance_ids, DryRun=True)
    except ClientError as e:
        if 'DryRunOperation' not in str(e):
            raise

    # Dry run succeeded, run start_instances without dryrun
    try:
        ec2_clients[region_name].start_instances(InstanceIds=instance_ids, DryRun=False)
    except ClientError as e:
        print(e)

    # Wait until instances are started
    waiter = ec2_clients[region_name].get_waiter('instance_running')
    waiter.wait(InstanceIds=instance_ids)


def start_all_instances(name_tag):
    instance_ids = get_all_instance_ids(name_tag)
    if not confirm_instances(instance_ids, "These instances will be started."):
        print("Aborting.")
        return
    print("Starting...")

    run_concurrently(start_instances, {region_name: [ids] for region_name, ids in instance_ids.items()})
    print_instances(instance_ids)


def stop_instances(region_name, instance_ids):

    if len(instance_ids) == 0:
        return

    # Do a dryrun first to verify permissions
    try:
        ec2_clients[region_name].stop_instances(InstanceIds=instance_ids, DryRun=True)
    except ClientError as e:
        if 'DryRunOperation' not in str(e):
            raise

    # Dry run succeeded, run stop_instances without dryrun
    try:
        ec2_clients[region_name].stop_instances(InstanceIds=instance_ids, DryRun=False)
    except ClientError as e:
        print(e)

    waiter = ec2_clients[region_name].get_waiter('instance_stopped')
    waiter.wait(InstanceIds=instance_ids)


def stop_all_instances(name_tag):
    instance_ids = get_all_instance_ids(name_tag)
    if not confirm_instances(instance_ids, "These instances will be stopped."):
        print("Aborting.")
        return
    print("Stopping...")

    run_concurrently(stop_instances, {region_name: [ids] for region_name, ids in instance_ids.items()})
    print_instances(instance_ids)


def get_public_ips(region_name, instance_ids):
    if len(instance_ids) == 0:
        return ""

    result = ec2_clients[region_name].describe_instances(InstanceIds=instance_ids)
    ips = ""
    for res in result['Reservations']:
        for inst in res['Instances']:
            ip = get_public_ip(inst)
            if ip != "":
                ips += ip+"\n"
            else:
                print("The following instance does not have a public IP address:")
                print(get_instance_info(region_name, [inst['InstanceId']])[0])
                exit(1)
    return ips


def print_public_ips(name_tag, output_file_name):
    instance_ids = get_all_instance_ids(name_tag)
    results = run_concurrently(get_public_ips, {region_name: [ids] for region_name, ids in instance_ids.items()})

    ips = ""
    for result in results.values():
        ips += result

    if output_file_name == "-" or output_file_name == "":
        print(ips, end="")
    else:
        with open(output_file_name, "w") as f:
            f.write(ips)


args = sys.argv[1:]

if len(args) == 0:
    print("Need to specify name tag of instances to select.")
    exit(1)
else:
    name_tag = args.pop(0)

# If no region file has been specified as the first command, try to load the default one.
if len(args) == 0 or args[0] != "regions":
    if pathlib.Path(DEFAULT_REGION_FILE).exists():
        load_regions(DEFAULT_REGION_FILE)
    else:
        print("Region file not specified and default ("+DEFAULT_REGION_FILE+") does not exist.")
        exit(1)

while len(args) > 0:
    command = args.pop(0)

    if command == "regions":
        if len(args) > 0:
            load_regions(args.pop(0))
        else:
            print("need to specify region file name")
            exit(1)
    elif command == "create":
        if len(args) > 1:
            create_all_instances(name_tag, int(args.pop(0)), args.pop(0))
        else:
            print("need to specify number of instances and instance type to create")
            exit(1)
    elif command == "terminate":
        terminate_all_instances(name_tag)
    elif command == "start":
        start_all_instances(name_tag)
    elif command == "stop":
        stop_all_instances(name_tag)
    elif command == "public-ips":
        if len(args) > 0:
            print_public_ips(name_tag, args.pop(0))
        else:
            print_public_ips(name_tag, "-")
    elif command == "list":
        print_instances(get_all_instance_ids(name_tag))
    else:
        print("unknown command: "+command)
        exit(1)
