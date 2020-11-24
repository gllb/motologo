package motologo

import "bytes"

var MOTOLOGO_FILE_LIST = [...]string {"logo_boot", "logo_battery", "logo_unlocked",
	"logo_lowpower", "logo_charge", "ic_unlocked_bootloader",
	"arrows", "yes", "wififlash", "switchconsole", "qcom", "factory",
	"droid_operation", "bptools", "barcodes", "start", "restartbootloader",
	"recoverymode", "poweroff", "bootloaderlogs", "bg", "red_fastboot",
	"orange_continue", "yellow_continue", "Switch_tools_mode", "yellow",
	"orange", "red"}

var MOTOLOGO_FILE_LIST_LEN = len(MOTOLOGO_FILE_LIST)

type MotologoHeader struct { // 11 bytes
	Signature [9]byte
	ItemCount uint32
}

type Motologo struct { // unknown size (>1024 bytes)
	Header MotologoHeader
	MotorunMetaSet []MotorunMeta
	_ [115]byte // ??? 115 is arbitrary it only work with 28 MotorunMeta record
	MotorunSet []bytes.Buffer
}

type MotorunMeta struct { // 32 bytes
	Name [24]byte
	Offset uint32
	Size   uint32
}
