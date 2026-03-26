package dcc

import "fmt"

// FIXME: Cleanup
// typedef void (*ACK_CALLBACK)(int16_t result);
type ackCallbackFn func(uint16, error)

func (d *DCC) writeCVByteMain(loco uint16, cv uint32, bValue byte) {
	/*
	   //
	   // writeCVByteMain: Write a byte with PoM on main. This writes
	   // the 5 byte sized packet to implement this DCC function
	   //

	   	void DCC::writeCVByteMain(int loco, int cv, byte bValue)  {
	   	  byte b[5];
	   	  byte nB = 0;
	   	  if (loco > HIGHEST_SHORT_ADDR)
	   	    b[nB++] = highByte(loco) | 0xC0;    // convert train number into a two-byte address

	   	  b[nB++] = lowByte(loco);
	   	  b[nB++] = cv1(WRITE_BYTE_MAIN, cv); // any CV>1023 will become modulus(1024) due to bit-mask of 0x03
	   	  b[nB++] = cv2(cv);
	   	  b[nB++] = bValue;

	   	  DCCWaveform::mainTrack.schedulePacket(b, nB, 4);
	   	}
	*/
}

func (d *DCC) writeCVBitMain(loco uint16, cv uint32, bNum byte, bValue bool) {
	/*
	   //
	   // writeCVBitMain: Write a bit of a byte with PoM on main. This writes
	   // the 5 byte sized packet to implement this DCC function
	   //

	   	void DCC::writeCVBitMain(int loco, int cv, byte bNum, bool bValue)  {
	   	  byte b[5];
	   	  byte nB = 0;
	   	  bValue = bValue % 2;
	   	  bNum = bNum % 8;

	   	  if (loco > HIGHEST_SHORT_ADDR)
	   	    b[nB++] = highByte(loco) | 0xC0;    // convert train number into a two-byte address

	   	  b[nB++] = lowByte(loco);
	   	  b[nB++] = cv1(WRITE_BIT_MAIN, cv); // any CV>1023 will become modulus(1024) due to bit-mask of 0x03
	   	  b[nB++] = cv2(cv);
	   	  b[nB++] = WRITE_BIT | (bValue ? BIT_ON : BIT_OFF) | bNum;

	   	  DCCWaveform::mainTrack.schedulePacket(b, nB, 4);
	   	}
	*/
}

func (d *DCC) writeCVByte(cv uint32, bValue byte, callback ackCallbackFn) {
	// void  DCC::writeCVByte(int16_t cv, byte byteValue, ACK_CALLBACK callback)  {
	//   DCCACK::Setup(cv, byteValue,  WRITE_BYTE_PROG, callback);
	// }
}

func (d *DCC) writeCVBit(cv uint32, bitValue bool, callback ackCallbackFn) {
	//	void DCC::writeCVBit(int16_t cv, byte bitNum, bool bitValue, ACK_CALLBACK callback)  {
	//	  if (bitNum >= 8) callback(-1);
	//	  else DCCACK::Setup(cv, bitNum, bitValue?WRITE_BIT1_PROG:WRITE_BIT0_PROG, callback);
	//	}
}

func (d *DCC) verifyCVByte(cv uint32, bValue byte, callback ackCallbackFn) {
	//	void  DCC::verifyCVByte(int16_t cv, byte byteValue, ACK_CALLBACK callback)  {
	//	  DCCACK::Setup(cv, byteValue,  VERIFY_BYTE_PROG, callback);
	//	}
}

func (d *DCC) verifyCVBit(cv uint32, bitNum byte, bitValue bool, callback ackCallbackFn) {
	//	void DCC::verifyCVBit(int16_t cv, byte bitNum, bool bitValue, ACK_CALLBACK callback)  {
	//	  if (bitNum >= 8) callback(-1);
	//	  else DCCACK::Setup(cv, bitNum, bitValue?VERIFY_BIT1_PROG:VERIFY_BIT0_PROG, callback);
	//	}
}

func (d *DCC) readCVBit(cv uint32, bitNum byte, callback ackCallbackFn) {
	if bitNum >= 8 {
		callback(0, fmt.Errorf("invalid bit number: %d", bitNum))
	} else {
		// FIXME: Implement
		// DCCACK::Setup(cv, bitNum,READ_BIT_PROG, callback);
	}
}

func (d *DCC) readCV(cv uint32, callback ackCallbackFn) {
	// FIXME: Implement
	// DCCACK::Setup(cv, 0,READ_CV_PROG, callback);
}

func (d *DCC) getLocoId(callback ackCallbackFn) {
	// FIXME: Implement
	// DCCACK::Setup(0,0, LOCO_ID_PROG, callback);
}

func (d *DCC) setLocoId(id uint16, callback ackCallbackFn) {
	if id < 1 || id > 10239 {
		callback(0, fmt.Errorf("invalid loco id: %d", id))
		return
	}
	if id <= shortAddressMax {
		// FIXME: Implement
		// DCCACK::Setup(id, SHORT_LOCO_ID_PROG, callback);
	} else {
		// FIXME: Implement
		// DCCACK::Setup(id | 0xc000,LONG_LOCO_ID_PROG, callback);
	}
}

func (d *DCC) setConsistId(id uint16, reverse bool, callback ackCallbackFn) {
	if id < 1 || id > 10239 {
		callback(0, fmt.Errorf("invalid loco id: %d", id))
		return
	}
	var cv19, cv20 byte

	if id <= shortAddressMax {
		cv19 = byte(id)
	} else {
		cv19 = byte(id % 100)
		cv20 = byte(id / 100)
	}
	if reverse {
		cv19 |= 0x80
	}
	// FIXME: Implement
	_ = cv20
	// DCCACK::Setup((cv20<<8)|cv19, CONSIST_ID_PROG, callback);
}

/* FIXME: Implement
const ackOp FLASH WRITE_BIT0_PROG[] = {
     BASELINE,
     W0,WACK,
     V0, WACK,  // validate bit is 0
     ITC1,      // if acked, callback(1)
     CALLFAIL  // callback (-1)
};
const ackOp FLASH WRITE_BIT1_PROG[] = {
     BASELINE,
     W1,WACK,
     V1, WACK,  // validate bit is 1
     ITC1,      // if acked, callback(1)
     CALLFAIL  // callback (-1)
};

const ackOp FLASH VERIFY_BIT0_PROG[] = {
     BASELINE,
     V0, WACK,  // validate bit is 0
     ITC0,      // if acked, callback(0)
     V1, WACK,  // validate bit is 1
     ITC1,
     CALLFAIL  // callback (-1)
};
const ackOp FLASH VERIFY_BIT1_PROG[] = {
     BASELINE,
     V1, WACK,  // validate bit is 1
     ITC1,      // if acked, callback(1)
     V0, WACK,
     ITC0,
     CALLFAIL  // callback (-1)
};

const ackOp FLASH READ_BIT_PROG[] = {
     BASELINE,
     V1, WACK,  // validate bit is 1
     ITC1,      // if acked, callback(1)
     V0, WACK,  // validate bit is zero
     ITC0,      // if acked callback 0
     CALLFAIL       // bit not readable
     };

const ackOp FLASH WRITE_BYTE_PROG[] = {
      BASELINE,
      WB,WACK,ITC1,    // Write and callback(1) if ACK
      // handle decoders that dont ack a write
      VB,WACK,ITC1,    // validate byte and callback(1) if correct
      CALLFAIL        // callback (-1)
      };

const ackOp FLASH VERIFY_BYTE_PROG[] = {
      BASELINE,
      BIV,         // ackManagerByte initial value
      VB,WACK,     // validate byte
      ITCB,       // if ok callback value
      STARTMERGE,    //clear bit and byte values ready for merge pass
      // each bit is validated against 0 and the result inverted in MERGE
      // this is because there tend to be more zeros in cv values than ones.
      // There is no need for one validation as entire byte is validated at the end
      V0, WACK, MERGE,        // read and merge first tested bit (7)
      ITSKIP,                 // do small excursion if there was no ack
        SETBIT,(ackOp)7,
        V1, WACK, NAKFAIL,    // test if there is an ack on the inverse of this bit (7)
        SETBIT,(ackOp)6,      // and abort whole test if not else continue with bit (6)
      SKIPTARGET,
      V0, WACK, MERGE,        // read and merge second tested bit (6)
      V0, WACK, MERGE,        // read and merge third  tested bit (5) ...
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, ITCBV,  // verify merged byte and return it if acked ok - with retry report
      CALLFAIL };


const ackOp FLASH READ_CV_PROG[] = {
      BASELINE,
      STARTMERGE,    //clear bit and byte values ready for merge pass
      // each bit is validated against 0 and the result inverted in MERGE
      // this is because there tend to be more zeros in cv values than ones.
      // There is no need for one validation as entire byte is validated at the end
      V0, WACK, MERGE,        // read and merge first tested bit (7)
      ITSKIP,                 // do small excursion if there was no ack
        SETBIT,(ackOp)7,
        V1, WACK, NAKFAIL,    // test if there is an ack on the inverse of this bit (7)
        SETBIT,(ackOp)6,      // and abort whole test if not else continue with bit (6)
      SKIPTARGET,
      V0, WACK, MERGE,        // read and merge second tested bit (6)
      V0, WACK, MERGE,        // read and merge third  tested bit (5) ...
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, ITCB,  // verify merged byte and return it if acked ok
      CALLFAIL };          // verification failed


const ackOp FLASH LOCO_ID_PROG[] = {
      BASELINE,
      // first check cv20 for extended addressing
      SETCV, (ackOp)20,     // CV 19 is extended
      SETBYTE, (ackOp)0,
      VB, WACK, ITSKIP,     // skip past extended section if cv20 is zero
      // read cv20 and 19 and merge
      STARTMERGE,           // Setup to read cv 20
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, NAKSKIP, // bad read of cv20, assume its 0
      BAD20SKIP,     // detect invalid cv20 value and ignore
      STASHLOCOID,   // keep cv 20 until we have cv19 as well.
      SETCV, (ackOp)19,
      STARTMERGE,           // Setup to read cv 19
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, NAKFAIL,  // cant recover if cv 19 unreadable
      COMBINE1920,  // Combile byte with stash and callback
// end of advanced 20,19 check
      SKIPTARGET,
      SETCV, (ackOp)19,     // CV 19 is consist setting
      SETBYTE, (ackOp)0,
      VB, WACK, ITSKIP,     // ignore consist if cv19 is zero (no consist)
      SETBYTE, (ackOp)128,
      VB, WACK, ITSKIP,     // ignore consist if cv19 is 128 (no consist, direction bit set)
      STARTMERGE,           // Setup to read cv 19
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, ITCB7,  // return 7 bits only, No_ACK means CV19 not supported so ignore it

      SKIPTARGET,     // continue here if CV 19 is zero or fails all validation
      SETCV,(ackOp)29,
      SETBIT,(ackOp)5,
      V0, WACK, ITSKIP,  // Skip to SKIPTARGET if bit 5 of CV29 is zero

      // Long locoid
      SETCV, (ackOp)17,       // CV 17 is part of locoid
      STARTMERGE,
      V0, WACK, MERGE,  // read and merge bit 1 etc
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, NAKFAIL,  // verify merged byte and return -1 it if not acked ok
      STASHLOCOID,         // keep stashed cv 17 for later
      // Read 2nd part from CV 18
      SETCV, (ackOp)18,
      STARTMERGE,
      V0, WACK, MERGE,  // read and merge bit 1 etc
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, NAKFAIL,  // verify merged byte and return -1 it if not acked ok
      COMBINELOCOID,        // Combile byte with stash to make long locoid and callback

      // ITSKIP Skips to here if CV 29 bit 5 was zero. so read CV 1 and return that
      SKIPTARGET,
      SETCV, (ackOp)1,
      STARTMERGE,
      SETBIT, (ackOp)6,  // skip over first bit as we know its a zero
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      V0, WACK, MERGE,
      VB, WACK, ITCB,  // verify merged byte and callback
      CALLFAIL
      };

const ackOp FLASH SHORT_LOCO_ID_PROG[] = {
      BASELINE,
      // Clear consist CV 19,20
      SETCV,(ackOp)20,
      SETBYTE, (ackOp)0,
      WB,WACK,     // ignore dedcoder without cv20 support
      SETCV,(ackOp)19,
      SETBYTE, (ackOp)0,
      WB,WACK,     // ignore dedcoder without cv19 support
      // Turn off long address flag
      SETCV,(ackOp)29,
      SETBIT,(ackOp)5,
      W0,WACK,
      V0,WACK,NAKFAIL,
      SETCV, (ackOp)1,
      SETBYTEL,   // low byte of word
      WB,WACK,ITC1,   // If ACK, we are done - callback(1) means Ok
      VB,WACK,ITC1,   // Some decoders do not ack and need verify
      CALLFAIL
};

// for CONSIST_ID_PROG the 20,19 values are already calculated
const ackOp FLASH CONSIST_ID_PROG[] = {
      BASELINE,
      SETCV,(ackOp)20,
      SETBYTEH,    // high byte to CV 20
      WB,WACK,ITSKIP,
      FAIL_IF_NONZERO_NAK, // fail if writing long address to decoder that cant support it
      SKIPTARGET,
      SETCV,(ackOp)19,
      SETBYTEL,   // low byte of word
      WB,WACK,ITC1,   // If ACK, we are done - callback(1) means Ok
      VB,WACK,ITC1,   // Some decoders do not ack and need verify
      CALLFAIL
};

const ackOp FLASH LONG_LOCO_ID_PROG[] = {
      BASELINE,
      // Clear consist CV 19,20
      SETCV,(ackOp)20,
      SETBYTE, (ackOp)0,
      WB,WACK,     // ignore dedcoder without cv20 support
      SETCV,(ackOp)19,
      SETBYTE, (ackOp)0,
      WB,WACK,     // ignore decoder without cv19 support
      // Turn on long address flag cv29 bit 5
      SETCV,(ackOp)29,
      SETBIT,(ackOp)5,
      W1,WACK,
      V1,WACK,NAKFAIL,
      // Store high byte of address in cv 17
      SETCV, (ackOp)17,
      SETBYTEH, // high byte of word
      WB,WACK,      // do write
      ITSKIP,       // if ACK, jump to SKIPTARGET
        VB,WACK,    // try verify instead
        ITSKIP,     // if ACK, jump to SKIPTARGET
          CALLFAIL, // if still here, fail
      SKIPTARGET,
      // store
      SETCV, (ackOp)18,
      SETBYTEL,   // low byte of word
      WB,WACK,ITC1,   // If ACK, we are done - callback(1) means Ok
      VB,WACK,ITC1,   // Some decoders do not ack and need verify
      CALLFAIL
};
*/
