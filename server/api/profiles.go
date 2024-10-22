package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"howett.net/plist"
)

func GeneralPayload(payloadType, payloadIdentifier, payloadDisplayName, payloadDescription, payloadOrganization string) map[string]any {
	generalMap := make(map[string]any)
	generalMap["PayloadType"] = payloadType
	generalMap["PayloadVersion"] = 1
	generalMap["PayloadIdentifier"] = payloadIdentifier
	generalMap["PayloadUUID"] = uuid.New().String()
	generalMap["PayloadDisplayName"] = payloadDisplayName
	generalMap["PayloadDescription"] = payloadDescription
	generalMap["PayloadOrganization"] = payloadOrganization
	return generalMap
}

func GetIPPort(r *http.Request) (string, string, error) {
	ip, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ip, port, err
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return ip, port, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}
	return ip, port, nil
}

func ProfileServicePayload(r *http.Request, challenge string) ([]byte, error) {
	sanitized := ""
	for _, id := range strings.Split(r.URL.Query().Get("profile-ids"), ",") {
		var i int
		for index, c := range id {
			if c < '0' || c > '9' || index > 15 {
				break
			}
			i = index
		}
		sanitized += id[:(i+1)] + ","
	}
	if len(sanitized) > 1 {
		sanitized = sanitized[:len(sanitized)-1]
	}

	queryString := fmt.Sprintf("SELECT company_code FROM profiles WHERE profile_id IN (%s)", sanitized)
	query, err := GetDB().Query(queryString)
	if err != nil {
		return []byte{}, err
	}

	displayName := "WRAPS Profile ("
	var value string
	for query.Next() {
		query.Scan(&value)
	}

	displayName = strings.TrimSuffix(displayName, ", ")
	displayName += ")"

	payload := GeneralPayload(
		"Profile Service",
		"com.wraps.mobileconfig.profile-service",
		displayName,
		"Install this profile to enroll in automatic wifi connection to the specified locations.",
		"com.wraps.wraps",
	)

	payload_content := make(map[string]any)
	ip, _, err := GetIPPort(r)
	if err != nil {
		return []byte{}, err
	}
	payload_content["URL"] = "https://" + ip + "/profile"
	payload_content["DeviceAttributes"] = []string{
		"UDID",
		"VERSION",
		"PRODUCT", // e.g. iPhone1,1 or iPod2,1
		"SERIAL",  // The device's serial number
		"MEID",    // The device's Mobile Equipment Identifier
		"IMEI",
	}

	if len(challenge) > 0 {
		payload_content["Challenge"] = challenge
	}

	payload["PayloadContent"] = payload_content
	pList, err := plist.Marshal(payload, plist.BinaryFormat)
	if err != nil {
		return []byte{}, err
	}

	return pList, nil
}
