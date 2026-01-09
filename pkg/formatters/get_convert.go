// © 2022 Nokia.
//
// This code is a Contribution to the gNMIc project ("Work") made under the Google Software Grant and Corporate Contributor License Agreement ("CLA") and governed by the Apache License 2.0.
// No other rights or licenses in or to any of Nokia's intellectual property are granted for any other purpose.
// This code is provided on an "as is" basis without any warranties of any kind.
//
// SPDX-License-Identifier: Apache-2.0

package formatters

import (
	"github.com/openconfig/gnmi/proto/gnmi"
)

// GetResponseToSubscribeResponses converts a GetResponse into a slice of SubscribeResponses.
// Each Notification in the GetResponse becomes a separate SubscribeResponse.
// This allows reusing the existing output infrastructure that expects SubscribeResponse messages.
func GetResponseToSubscribeResponses(rsp *gnmi.GetResponse) []*gnmi.SubscribeResponse {
	if rsp == nil {
		return nil
	}
	notifications := rsp.GetNotification()
	responses := make([]*gnmi.SubscribeResponse, 0, len(notifications))
	for _, notif := range notifications {
		sr := &gnmi.SubscribeResponse{
			Response: &gnmi.SubscribeResponse_Update{
				Update: notif,
			},
		}
		responses = append(responses, sr)
	}
	return responses
}

// GetResponseNotificationCount returns the number of notifications in a GetResponse.
func GetResponseNotificationCount(rsp *gnmi.GetResponse) int {
	if rsp == nil {
		return 0
	}
	return len(rsp.GetNotification())
}

// GetResponseUpdateCount returns the total number of updates across all notifications in a GetResponse.
func GetResponseUpdateCount(rsp *gnmi.GetResponse) int {
	if rsp == nil {
		return 0
	}
	count := 0
	for _, notif := range rsp.GetNotification() {
		count += len(notif.GetUpdate())
	}
	return count
}
