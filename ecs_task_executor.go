package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/fatih/structs"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
	"time"
	"reflect"
)

type Options struct {
	Cluster string `long:"cluster" description:"A name of ECS Cluster"`
	TaskDef string `short:"t" long:"taskdef" description:"A name of target of taskdefinition"`
	Name    string `short:"n" long:"name" description:"A name of target container name"`
	Command string `short:"c" long:"command" description:"Command which override default one"`
	Timeout int    `long:"timeout" description:"Timeout Value, defualt 600[sec]" default:"600"`
}

func main() {
	// Define options
	var opts Options
	var parser = flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := validate(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmtCommand := formatCommand(opts.Command)

	// Define task input. Override by args.
	svc := ecs.New(session.New(), &aws.Config{Region: aws.String(os.Getenv("AWS_DEFAULT_REGION"))})
	runInput := &ecs.RunTaskInput{
		Cluster:        aws.String(opts.Cluster),
		TaskDefinition: aws.String(opts.TaskDef),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Name:    aws.String(opts.Name),
					Command: fmtCommand,
				},
			},
		},
	}

	// Run Task
	runResult, err := svc.RunTask(runInput)
	if err != nil {
		parseRunTaskErr(err)
		os.Exit(1)
	}

	// Define task describe input
	taskArn := *runResult.Tasks[0].TaskArn
	describeInput := &ecs.DescribeTasksInput{
		Cluster: aws.String(opts.Cluster),
		Tasks: []*string{
			aws.String(taskArn),
		},
	}

	// Task status check loop
	startAt := time.Now()
	endAt := startAt.Add(time.Duration(opts.Timeout) * time.Second)
	for endAt.After(time.Now()) {
		describeResult, err := svc.DescribeTasks(describeInput)
		if err != nil {
			parseDescribeTaskErr(err)
			os.Exit(1)
		}
		if len(describeResult.Failures) > 0 {
			fmt.Println("Failed to get TaskDefinition: %s. Retrying...\n", opts.TaskDef)
		} else {
			checkStatus(describeResult, startAt)
		}
		time.Sleep(5 * time.Second)
		fmt.Printf("LastStatus=%s TimeElapsed=%s\n", *describeResult.Tasks[0].LastStatus, time.Now().Sub(startAt))
	}

	// Error. Exit after timeout
	fmt.Println("Timeout. Please check logs or extend timeout value.")
	os.Exit(1)
}

func checkStatus(t *ecs.DescribeTasksOutput, startAt time.Time) {
	// TODO: consider multipul containers or tasks case.
	task := *t.Tasks[0]
	con := task.Containers[0]
	// Continue if task is not stopped
	if *task.LastStatus != "STOPPED" {
		return
	}

	// Task status check after stopped
	// Error: ExitCode is not defined
	// Error: ExitCode is not 0
	// Success: ExitCode is 0
	mapCon := structs.Map(*con)
	if reflect.ValueOf(mapCon["ExitCode"]).IsNil() {
		fmt.Println(con)
		os.Exit(1)
	} else {
		if *con.ExitCode == 0 {
			fmt.Printf("LastStatus=%s TimeElapsed=%s\n", *task.LastStatus, time.Now().Sub(startAt))
			fmt.Println("Task successfully finished.")
		} else {
			fmt.Println(con)
		}
		os.Exit(int(*con.ExitCode))
	}
}

func validate(opts Options) error {
	if opts.Cluster == "" {
		return errors.New("[OPTION VALIDATION ERROR] --cluster option is required.")
	}
	if opts.TaskDef == "" {
		return errors.New("[OPTION VALIDATION ERROR] -t option is required.")
	}
	if opts.Name == "" {
		return errors.New("[OPTION VALIDATION ERROR] -n option is required.")
	}
	if opts.Command == "" {
		return errors.New("[OPTION VALIDATION ERROR] -c option is required.")
	}
	fmt.Printf("Set timeout as %d sec.\n", opts.Timeout)
	return nil
}

func formatCommand(cmd string) []*string {
	fmtCommand := []*string{}
	splitCmd := strings.Split(cmd, " ")
	for _, c := range splitCmd {
		fmtCommand = append(fmtCommand, aws.String(c))
	}
	return fmtCommand
}

func parseRunTaskErr(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case ecs.ErrCodeServerException:
			fmt.Println(ecs.ErrCodeServerException, aerr.Error())
		case ecs.ErrCodeClientException:
			fmt.Println(ecs.ErrCodeClientException, aerr.Error())
		case ecs.ErrCodeInvalidParameterException:
			fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
		case ecs.ErrCodeClusterNotFoundException:
			fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
		case ecs.ErrCodeUnsupportedFeatureException:
			fmt.Println(ecs.ErrCodeUnsupportedFeatureException, aerr.Error())
		case ecs.ErrCodePlatformUnknownException:
			fmt.Println(ecs.ErrCodePlatformUnknownException, aerr.Error())
		case ecs.ErrCodePlatformTaskDefinitionIncompatibilityException:
			fmt.Println(ecs.ErrCodePlatformTaskDefinitionIncompatibilityException, aerr.Error())
		case ecs.ErrCodeAccessDeniedException:
			fmt.Println(ecs.ErrCodeAccessDeniedException, aerr.Error())
		case ecs.ErrCodeBlockedException:
			fmt.Println(ecs.ErrCodeBlockedException, aerr.Error())
		default:
			fmt.Println(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	}
}

func parseDescribeTaskErr(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case ecs.ErrCodeServerException:
			fmt.Println(ecs.ErrCodeServerException, aerr.Error())
		case ecs.ErrCodeClientException:
			fmt.Println(ecs.ErrCodeClientException, aerr.Error())
		case ecs.ErrCodeInvalidParameterException:
			fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
		case ecs.ErrCodeClusterNotFoundException:
			fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
		default:
			fmt.Println(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
	}
	return
}

