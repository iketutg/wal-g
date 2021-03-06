package walparser

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	// info flags

	XlrInfoMask     = 0x0F
	XlrRmgrInfoMask = 0xF0

	XlrSpecialRelUpdate  = 0x01
	XlrCheckConsistency  = 0x02
	XLogRecordHeaderSize = 24
)

type InconsistentXLogRecordTotalLengthError struct {
	totalRecordLength uint32
}

func (err InconsistentXLogRecordTotalLengthError) Error() string {
	return fmt.Sprintf("total record length is too small: %v, expected at least: %v", err.totalRecordLength, XLogRecordHeaderSize)
}

type InvalidXLogRecordResourceManagerIDError struct {
	resourceManagerID uint8
}

func (err InvalidXLogRecordResourceManagerIDError) Error() string {
	return fmt.Sprintf("resource manager id is invalid: %v, while it should be less then: %v", err.resourceManagerID, RmNextFreeID)
}

var ZeroRecordHeaderError = errors.New("whole record header is zero, maybe it's parsed from .partial file or after WAL-Switch operation")

/* This struct corresponds to postgres struct XLogRecord.
 * For clarification you can find it in postgres:
 * src/include/access/xlogrecord.h
 */
type XLogRecordHeader struct {
	TotalRecordLength uint32
	XactID            uint32
	PrevRecordPtr     XLogRecordPtr
	Info              uint8
	ResourceManagerID uint8
	/* 2 bytes of padding here, initialize to zero */
	Crc32Hash uint32
	/* XLogRecordBlockHeaders and XLogRecordDataHeader follow, no padding */
}

func (header *XLogRecordHeader) checkTotalRecordLengthConsistency() error {
	if header.TotalRecordLength < XLogRecordHeaderSize {
		return InconsistentXLogRecordTotalLengthError{header.TotalRecordLength}
	}
	return nil
}

func (header *XLogRecordHeader) checkResourceManagerIDValidity() error {
	if header.ResourceManagerID >= RmNextFreeID {
		return InvalidXLogRecordResourceManagerIDError{header.ResourceManagerID}
	}
	return nil
}

func (header *XLogRecordHeader) checkConsistency() error {
	err := header.checkTotalRecordLengthConsistency()
	if err != nil {
		if header.isZero() {
			return ZeroRecordHeaderError
		}
		return err
	}
	return header.checkResourceManagerIDValidity()
}

func (header *XLogRecordHeader) isZero() bool {
	return header.TotalRecordLength == 0 &&
		header.XactID == 0 &&
		header.PrevRecordPtr == 0 &&
		header.Info == 0 &&
		header.ResourceManagerID == 0 &&
		header.Crc32Hash == 0
}
