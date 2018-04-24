# Jenkins Pipeline Tutorial

In this tutorial you will create a simple CI/CD pipeline using Jenkins, the [Pipeline plugin][1]
and [AWS Fargate][2].

## Table of Contents

- [Jenkins Pipeline Tutorial](#jenkins-pipeline-tutorial)
    - [Table of Contents](#table-of-contents)
    - [Prerequisites](#prerequisites)
    - [Forking the Repository](#forking-the-repository)
    - [Preparing the Deployment Environment](#preparing-the-deployment-environment)
    - [Prepare a Docker Repository on ECR](#prepare-a-docker-repository-on-ecr)
    - [Preparing Jenkins](#preparing-jenkins)
        - [Create an IAM Role for Jenkins](#create-an-iam-role-for-jenkins)
        - [Create an EC2 Instance](#create-an-ec2-instance)
        - [Install Jenkins](#install-jenkins)
        - [Install Git and Docker on the Jenkins Instance](#install-git-and-docker-on-the-jenkins-instance)
        - [Run the Jenkins Setup Wizard](#run-the-jenkins-setup-wizard)
    - [Creating the Pipeline](#creating-the-pipeline)
    - [Configuring the Pipeline](#configuring-the-pipeline)
        - [Looking at the Sample Pipeline](#looking-at-the-sample-pipeline)
        - [Running the Pipeline](#running-the-pipeline)
        - [Adding a CI Stage](#adding-a-ci-stage)
        - [Adding a CD Stage](#adding-a-cd-stage)
    - [Testing the Pipeline](#testing-the-pipeline)
    - [Cleaning Up](#cleaning-up)

## Prerequisites

You will need the following to complete this tutorial:

- An AWS account
- [AWS CLI][3] installed and configured for your AWS account
- [Git][8]

## Forking the Repository

You will have to push changes to Github in order to trigger the CI/CD pipeline. Therefore, before
going any further in this tutorial, **fork this repository** and work on your own fork from now on.
If you have never forked a repository, [this][10] might help.

## Preparing the Deployment Environment

The simplest way to get a container running on AWS is to use Fargate, so we will use a sample
Fargate deployment which takes care of everything for us: VPC, SGs, IAM and the cluster itself.

> **NOTE:** Fargate is only available on the **us-east-1** region at the time of writing, so we
> will use this region.

Perform the following steps to prepare the Fargate environment for our deployment:

1. Log in to the ECS console under the **us-east-1** region (or simply click [here][4]).
2. Click **Get Started**.
3. Under **Container definition**, leave the **sample-app** option selected and click **Next**.
4. Under **Load balancer type** choose **Application Load Balancer** and click **Next**.
5. If you want, change the name of the cluster under **Cluster name**. Click **Next**.
6. Click **Create**.

The cluster should now be created along with all required resources. A sample app will be
automatically deployed on the cluster once created (this could take a few minutes).

Verify that the sample app works by browsing the DNS name of the load balancer. To find the DNS
name you can click the link near **Load balancer** in the cluster creation status page, or find the
load balancer in the [Load Balancers view][5].

## Prepare a Docker Repository on ECR

The pipeline you are about to create will generate Docker images, which need to be pushed to some
Docker registry. Since we are on AWS we can easily use ECR for this purpose, however you may use
other Docker registries as well.

To create a repository on ECR, follow these steps:

1. On the [Repositories][14] section on the ECS console, click **Get started** or **Create
repository**.
2. Under **Repository name** type "sample-app" and click **Next step** then **Done**.

Make note of the **Repository URI** - you will need it later.

## Preparing Jenkins

You will need a Jenkins instance for this tutorial. Perform the following steps to deploy Jenkins
inside your AWS account:

### Create an IAM Role for Jenkins

1. In the [IAM roles view][6] click **Create role**.
2. Choose **EC2** and click **Next: Permissions**.
3. Check the **AmazonECS_FullAccess** and the **AmazonEC2ContainerRegistryPowerUser** policies and
click **Next: Review**.
4. Under **Role name** type "Jenkins" and click **Create role**.

### Create an EC2 Instance

> **NOTE:** If you are lazy, you can use a pre-configured AMI with Jenkins, Git and Docker. To do so,
> use **jenkins-docker 1524155485 (ami-fc47eb83)** in step 2 and jump directly to
> [Run the Jenkins Setup Wizard](#run-the-jenkins-setup-wizard) after launching the instance.

1. In the [EC2 console][7] click **Launch Instance**.
2. Select an **Amazon Linux** AMI.
3. Leave **t2.micro** selected and click **Next: Configure Instance Details**.
4. Under **Network**, choose any VPC with a public subnet. The **default VPC** will work fine here,
too.
5. Under **Subnet**, choose any public subnet.
6. Under **Auto-assign Public IP** choose **Enabled**.
7. Under **IAM role** choose the **Jenkins** role you created before.
8. Click **Next: Add Storage**.
9. Click **Next: Add Tags**.
10. Create a tag with the key "Name" and the value "Jenkins" and click **Next: Configure Security Group**.
11. Name the security group "Jenkins" and allow **SSH access** as well as access to **TCP port
8080** from your WAN IP address.
12. Click **Review and Launch**.
13. Click **Launch**.
14. Choose an existing SSH key or create a new one, then click **Launch Instances**.

### Install Jenkins

> **Note:** If you've deployed Jenkins from a pre-configured AMI, go directly to
> [Run the Jenkins Setup Wizard](#run-the-jenkins-setup-wizard).

1. SSH into the instance you created in the previous step.
2. Run the following commands to install Jenkins:

        sudo yum remove -y java
        sudo yum install -y java-1.8.0-openjdk
        sudo wget -O /etc/yum.repos.d/jenkins.repo http://pkg.jenkins-ci.org/redhat-stable/jenkins.repo
        sudo rpm --import https://jenkins-ci.org/redhat/jenkins-ci.org.key
        sudo yum install -y jenkins

3. Run `sudo service jenkins start` to start the Jenkins service.

### Install Git and Docker on the Jenkins Instance

1. SSH into the Jenkins instance if you are not already there.
2. Run the following commands on the Jenkins instance:

        sudo yum install -y git docker
        sudo service docker start
        # Allow Jenkins to talk to the Docker daemon
        sudo usermod -aG docker jenkins
        sudo service jenkins restart

### Run the Jenkins Setup Wizard

1. Browse `http://<instance_ip>:8080`.
2. Under **Administrator password** enter the output of `sudo cat /var/lib/jenkins/secrets/initialAdminPassword` on the Jenkins instance.
3. Click **Continue**.
4. Click **Install suggested plugins** and let the installation finish.
5. Click **Continue as admin**.
6. Click **Start using Jenkins**.

## Creating the Pipeline

We will now create the Jenkins pipeline. Perform the following steps on the Jenkins UI:

1. Click **New Item** to create a new Jenkins job.
2. Under **Enter an item name** type "sample-pipeline".
3. Choose **Pipeline** as the job type and click **OK**.
4. Under **Pipeline -> Definition** choose **Pipeline script from SCM**.
5. Under **SCM** choose **Git**.
6. Under **Repository URL** paste the [HTTPS URL][9] of your (forked) repository.

> **NOTE:** It is generally recommended to use Git **over SSH** rather than HTTPS, especially in
> automated processes. However, to simplify things and since the repository is public, we can
> simply use the HTTPS URL instead of dealing with SSH keys.

7. Leave the rest at the default and click **Save**.

You should now have a pipeline configured. When executing the pipeline, Jenkins will clone the Git
repository, look for a file named `Jenkinsfile` at its root and execute the instructions in it.

## Configuring the Pipeline

The Jenkins Pipeline plugin supports two types of piplines: **declarative pipelines** and
**scripted pipelines**. Declarative pipelines are simpler and have a nice, clean syntax. Scripted
pipelines, on the other hand, offer unlimited flexibility by exposing the full power of the
[Groovy][11] programming language in the pipeline.

In this tutorial we will use the **declarative syntax**, which is more than enough for what we are
trying to accomplish.

### Looking at the Sample Pipeline

Let's take a look at the sample pipeline that is already in the repository. Open the file called
`Jenkinsfile` file in a text editor (preferably one [which][12] [supports][13] the Jenkinsfile
syntax).

We can see that the entire pipeline is inside a top-level directive called `pipeline`.

Then we have a line saying `agent any` - this is required for declarative pipelines, but we are not
going to touch it in this tutorial. If you are still curious about what the agent directive does,
you can read about it [here][15].

Next we have the `environment` directive. This section allows us to configure global variables
which will be available (for both reading and writing) in any of the pipeline's stages. This is
useful for configuring global settings.

Lastly, we have a `stage`. You can have as many stages as you want in a pipeline. A stage is a
major section of the pipeline and it contains the actual "work" which the pipeline does. This work
is defined in `steps`. A step can execute a shell script, push an artifact somewhere, send an email
or a Slack message to someone and do lots of other stuff. We can see that at the moment our
pipeline doesn't do much, just prints something to the console using an `echo` step.

> **NOTE:** There is [an entire list][16] of step types which can be used in Jenknis pipelines,
> however in this tutorial we will keep things simple and use mostly the `sh` step, which executes
> a shell script.

So, now that we understand the structure of our pipeline, let's run it.

### Running the Pipeline

1. From the top-level view on the Jenkins UI, click on the pipeline's name ("sample-pipeline").
2. On the menu to the left, click **Build Now**.

This will trigger a run. You should see a new run (or "build") under the **Build History** view on
on the left side. To see the logs from the build, click the build number (`#1` if this is your
first build) and the click **Console Output**.

If all went well, after some Git-related output you should see that the pipeline ran the only stage
we currently have, which should simply print `This is a sample stage`.

Great. Now let's make the pipeline do some real stuff.

### Adding a CI Stage

Let's add a simple CI step to our pipeline. We want to build a Docker image from our app and push
it to ECR so that we can later deploy containers from it.

Let's populate the `docker_repo_uri` environment variable with the full URI of the ECR repository
you created previously. It shall be similar to the following:

    pipeline {
        ...

        environment {
            region = "us-east-1"
            docker_repo_uri = "xxxxxxxxxxxx.dkr.ecr.us-east-1.amazonaws.com/sample-app"
            task_def_arn = ""
            cluster = ""
            exec_role_arn = ""
        }

        ...
    }

Now, replace the "Example" stage with the following:

    stage('Build') {
        steps {
            // Get SHA1 of current commit
            script {
                commit_id = sh(script: "git rev-parse --short HEAD", returnStdout: true).trim()
            }
            // Build the Docker image
            sh "docker build -t ${docker_repo_uri}:${commit_id} ."
            // Get Docker login credentials for ECR
            sh "aws ecr get-login --no-include-email --region ${region} | sh"
            // Push Docker image
            sh "docker push ${docker_repo_uri}:${commit_id}"
            // Clean up
            sh "docker rmi -f ${docker_repo_uri}:${commit_id}"
        }
    }

Notice that we have two types of steps here: `script` and `sh`. `script` steps allow us to run a
Groovy code snippet inside our declarative pipeline. We need this because we want to capture the
SHA1 of the current commit and assign it to a variable, which we can then use to uniquely tag the
Docker image we are building. `sh` steps are simply shell commands.

So now our pipeline should build a Docker image, push it to ECR and clean up the leftover image so
that we don't accumulate garbage on Jenkins.

In order to update the pipeline, we must commit and push our changes to Github. So, when you are
done editing, do the following:

1. Commit your changes by running `git commit -am "Add CI step to pipeline"`.
2. Push your changes to Github by running `git push origin`.

Now, re-run the pipeline on Jenkins and examine its output. If all goes well, the pipeline will
build a Docker image, push it to ECR and clean up the local image on Jenkins. Verify this by
looking at the [Repositories][14] section of the ECS console. Your repository should now have an
image in it.

### Adding a CD Stage

Now that we have a pipeline which automatically generates Docker images for us, let's add another
stage that will deploy new images to our deployment environment (Fargate).

In your editor, open the `taskdef.json` file. This file defines how to deploy our app to Fargate.
It already contains everything we need except one thing: the Docker image to use for the
deployment. As you can see, at the moment the `image` key contains a placeholder - `{{image}}`.
This placeholder is **invalid** if you just try to submit it as-is to Fargate, and must be edited
first. However, we can't simply hardcode a specific Docker image here, because we want our pipeline
to every time deploy **the image we just built** to the environment. So, we will override this
field with the correct value in our new stage.

Before adding the stage, populate the following variables in your `environment`:

`task_def_arn` - should contain the ARN of the task definition Fargate has already created for
you, **without the revision number** (`:1` etc.).

> **Note:** You can look for this ARN under the [Task Definitions view][18] on the ECS console, or
> by running `aws ecs list-task-definitions | grep first-run-task-definition`.

`cluster` - should contain the name of the Fargate cluster you created before.

`exec_role_arn` - should contain the ARN of the **ecsTaskExecutionRole** role which was created
automatically for you when you created the cluster.

> **Note:** In case you don't have such a role, you can create it using [these instructions][17].

So, after these changes, your `environment` should look similar to the following:

    environment {
        region = "us-east-1"
        docker_repo_uri = "xxxxxxxxxxxx.dkr.ecr.us-east-1.amazonaws.com/sample-app"
        task_def_arn = "arn:aws:ecs:us-east-1:xxxxxxxxxxxx:task-definition/first-run-task-definition"
        cluster = "default"
        exec_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/ecsTaskExecutionRole"
    }

Now, add the following stage right after the existing "Build" stage:

    stage('Deploy') {
        steps {
            // Override image field in taskdef file
            sh "sed -i 's|{{image}}|${docker_repo_uri}:${commit_id}|' taskdef.json"
            // Create a new task definition revision
            sh "aws ecs register-task-definition --execution-role-arn ${exec_role_arn} --cli-input-json file://taskdef.json --region ${region}"
            // Update service on Fargate
            sh "aws ecs update-service --cluster ${cluster} --service sample-app-service --task-definition ${task_def_arn} --region ${region}"
        }
    }

The first step in this stage overrides the `image` field in taskdef.json with the name of the image
that has been created in the CI stage.

> **Note:** We use `|` as a delimiter in `sed` because `${docker_repo_uri}` contains a slash, which
> creates escaping problems in this case.

The second step registers a new task definition revision which references the new image we already
have in ECR.

The last step instructs Fargate to update the app on the cluster, which will cause a new container
to be launched from the image we just pushed to ECR, replacing the old one.

> **Note:** When calling `update-service` you may specify a specific **task definition revision**
> by including the revision number in the provided ARN (for example `:3`). When not doing so,
> Fargate simply uses the most recent revision, which is fine in our case.

So, we should now be ready to test our CD stage. Commit your changes, push them to Github and run
the pipeline. If all goes well, an update should be triggered on Fargate, which will deploy our
app instead of the sample app we deployed using the Getting Started wizard. Verify this by browsing
the DNS name of the load balancer again. If you still see the sample app, the deployment might
still be in progress. You can follow it on the service's [Deployments][19] tab.

## Testing the Pipeline

Up to now, all we did was set up a CI/CD pipeline which will build and deploy code changes
automatically. Now, we will verify it actually does so by making a very simple code change.

Open `app.go` in your editor and change `version = "1.0"` on line 11 to `version = "1.1"`. Push
the change to Github and run the pipeline. If all goes well, after a short time you should see the
"Version" field change when refreshing your browser.

> **Note:** Deploying a new version could take a few minutes, mainly because the default
> [Deregistration Delay][20] is 5 minutes. You may reduce this timer to speed up deployments, or
> manually kill the old tasks.

## Cleaning Up

When you are done experimenting and would like to delete the environment, perform the following:

1. Terminate the Jenkins instance and delete its IAM role, security group and SSH key.
2. On the [Clusters view][21] of the ECS console, choose your cluster, click **Delete Cluster** and
then **Delete**. This will delete everything Fargate has created for you including the VPC and the
load balancer.
3. In the [Task Definitions view][18], click **first-run-task-definition**, check all of the
revisions in the list, then click **Actions -> Deregister** and then **Deregister**.
4. In the [Repositories view][14], check the repository you created, then click **Delete
repository** and then **Delete**.

[1]: https://jenkins.io/doc/book/pipeline/
[2]: https://aws.amazon.com/fargate/
[3]: https://aws.amazon.com/cli/
[4]: https://console.aws.amazon.com/ecs/home?region=us-east-1
[5]: https://console.aws.amazon.com/ec2/v2/home?region=us-east-1#LoadBalancers:sort=loadBalancerName
[6]: https://console.aws.amazon.com/iam/home?region=us-east-1#/roles
[7]: https://console.aws.amazon.com/ec2/v2/home?region=us-east-1
[8]: https://git-scm.com/
[9]: https://help.github.com/articles/which-remote-url-should-i-use/
[10]: https://guides.github.com/activities/forking/
[11]: http://groovy-lang.org/
[12]: https://code.visualstudio.com/
[13]: https://atom.io/
[14]: https://console.aws.amazon.com/ecs/home?region=us-east-1#/repositories
[15]: https://jenkins.io/doc/book/pipeline/syntax/#agent
[16]: https://jenkins.io/doc/pipeline/steps/
[17]: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html
[18]: https://console.aws.amazon.com/ecs/home?region=us-east-1#/taskDefinitions
[19]: https://console.aws.amazon.com/ecs/home?region=us-east-1#/clusters/default/services/sample-app-service/deployments
[20]: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/load-balancer-target-groups.html#deregistration-delay
[21]: https://console.aws.amazon.com/ecs/home?region=us-east-1#/clusters
