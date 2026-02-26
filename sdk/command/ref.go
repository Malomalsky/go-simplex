package command

import "strconv"

func DirectRef(contactID int64) string {
	return "@" + strconv.FormatInt(contactID, 10)
}

func GroupRef(groupID int64) string {
	return "#" + strconv.FormatInt(groupID, 10)
}

func LocalRef(folderID int64) string {
	return "*" + strconv.FormatInt(folderID, 10)
}
