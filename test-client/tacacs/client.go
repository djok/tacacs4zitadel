package tacacs

import (
	"fmt"
	"net"

	tq "github.com/facebookincubator/tacquito"
)

type Client struct {
	client *tq.Client
}

func NewClient(conn net.Conn, secret []byte) *Client {
	// Create a Tacquito client that will use the connection
	// We need to close the original connection and create a new one through Tacquito
	conn.Close()
	
	// Extract network and address from original connection
	remoteAddr := conn.RemoteAddr().String()
	network := conn.RemoteAddr().Network()
	
	client, err := tq.NewClient(tq.SetClientDialer(network, remoteAddr, secret))
	if err != nil {
		// Return a dummy client, error will be caught later
		return &Client{client: nil}
	}
	
	return &Client{client: client}
}

func (c *Client) Authenticate(username, password string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("client not initialized")
	}
	
	// Create PAP authentication packet
	packet := tq.NewPacket(
		tq.SetPacketHeader(
			tq.NewHeader(
				tq.SetHeaderVersion(tq.Version{MajorVersion: tq.MajorVersion, MinorVersion: tq.MinorVersionOne}),
				tq.SetHeaderType(tq.Authenticate),
				tq.SetHeaderRandomSessionID(),
			),
		),
		tq.SetPacketBodyUnsafe(
			tq.NewAuthenStart(
				tq.SetAuthenStartType(tq.AuthenTypePAP),
				tq.SetAuthenStartAction(tq.AuthenActionLogin),
				tq.SetAuthenStartPrivLvl(tq.PrivLvl(1)),
				tq.SetAuthenStartService(tq.AuthenServiceLogin),
				tq.SetAuthenStartUser(tq.AuthenUser(username)),
				tq.SetAuthenStartData(tq.AuthenData(password)),
			),
		),
	)

	// Send packet
	resp, err := c.client.Send(packet)
	if err != nil {
		return false, err
	}

	// Parse response
	var authenReply tq.AuthenReply
	if err := tq.Unmarshal(resp.Body, &authenReply); err != nil {
		return false, err
	}

	return authenReply.Status == tq.AuthenStatusPass, nil
}

func (c *Client) Authorize(username, command string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("client not initialized")
	}
	
	// Create authorization packet
	packet := tq.NewPacket(
		tq.SetPacketHeader(
			tq.NewHeader(
				tq.SetHeaderVersion(tq.Version{MajorVersion: tq.MajorVersion, MinorVersion: tq.MinorVersionOne}),
				tq.SetHeaderType(tq.Authorize),
				tq.SetHeaderRandomSessionID(),
			),
		),
		tq.SetPacketBodyUnsafe(
			tq.NewAuthorRequest(
				tq.SetAuthorRequestMethod(tq.AuthenMethodNone),
				tq.SetAuthorRequestPrivLvl(tq.PrivLvl(1)),
				tq.SetAuthorRequestType(tq.AuthenTypeASCII),
				tq.SetAuthorRequestService(tq.AuthenServiceLogin),
				tq.SetAuthorRequestUser(tq.AuthenUser(username)),
				tq.SetAuthorRequestArgs(tq.Args{tq.Arg(command)}),
			),
		),
	)

	// Send packet
	resp, err := c.client.Send(packet)
	if err != nil {
		return false, err
	}

	// Parse response
	var authorReply tq.AuthorReply
	if err := tq.Unmarshal(resp.Body, &authorReply); err != nil {
		return false, err
	}

	return authorReply.Status == tq.AuthorStatusPassAdd || authorReply.Status == tq.AuthorStatusPassRepl, nil
}

func (c *Client) Account(username, command string) error {
	if c.client == nil {
		return fmt.Errorf("client not initialized")
	}
	
	// Create accounting packet
	packet := tq.NewPacket(
		tq.SetPacketHeader(
			tq.NewHeader(
				tq.SetHeaderVersion(tq.Version{MajorVersion: tq.MajorVersion, MinorVersion: tq.MinorVersionOne}),
				tq.SetHeaderType(tq.Accounting),
				tq.SetHeaderRandomSessionID(),
			),
		),
		tq.SetPacketBodyUnsafe(
			tq.NewAcctRequest(
				tq.SetAcctRequestFlag(tq.AcctFlagStart),
				tq.SetAcctRequestMethod(tq.AuthenMethodNone),
				tq.SetAcctRequestPrivLvl(tq.PrivLvl(1)),
				tq.SetAcctRequestType(tq.AuthenTypeASCII),
				tq.SetAcctRequestService(tq.AuthenServiceLogin),
				tq.SetAcctRequestUser(tq.AuthenUser(username)),
				tq.SetAcctRequestArgs(tq.Args{tq.Arg(command)}),
			),
		),
	)

	// Send packet
	resp, err := c.client.Send(packet)
	if err != nil {
		return err
	}

	// Parse response
	var acctReply tq.AcctReply
	if err := tq.Unmarshal(resp.Body, &acctReply); err != nil {
		return err
	}

	if acctReply.Status != tq.AcctReplyStatusSuccess {
		return fmt.Errorf("accounting failed with status: %v", acctReply.Status)
	}

	return nil
}