# autoparter2
Automatically preparing partitions and mount to specific dirs, based on AWS EBS volume tag:

Just specify TAG `mount` for each volume which you want to mount.
```
Key=mount
Value=/directory/path
```
Also add UUID params to ```/etc/fstab```.
I you need to stop particular service which can interfere mount/umount process, please use flag `-s` with particallar service name:
Defaults: FS type `xfs`, if you want to change use option `-f ext4`.
## Minimum requirements for EC2 Instance profile:
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "ec2:DescribeVolumes",
            "Resource": "*"
        }
    ]
}
```
### Tested on:
Ubuntu 18.04 `ami-0f2b111fdc1647918`

Ubuntu 20.04 `ami-0ea142bd244023692`

AMZ Linux2   `ami-007a607c4abd192db`

RHT8         `ami-029ba835ddd43c34f`

#### t2,t3,m5,m4,c5 Instances type.

### Examples:
```
./autopart -s "cron,sshd,apache2"
./autopart -f ext4 -s "apache2"
./autopart -f ext4
./autopart
```

Based on Golang 1.15
Specifically designed for AWS EC2 instances

### Yuriy Yurov.
