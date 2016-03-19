package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"sync"
	"log"
	rtl "github.com/jpoirier/gortlsdr"

)

type UAT struct {
	dev *rtl.Context
	wg  *sync.WaitGroup
}

func (u *UAT) sigAbort() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	u.shutdown()
	os.Exit(0)
}

func (u *UAT) shutdown() {
	fmt.Println()
	log.Println("UAT shutdown(): closing device ...")
	u.dev.Close() // preempt the blocking ReadSync call
	log.Println("UAT shutdown(): calling uatWG.Wait() ...")
	u.wg.Wait() // Wait for the goroutine to shutdown
}

func (u *UAT) read() {
	defer u.wg.Done()
	log.Println("Entered UAT read() ...")

	var readCnt uint64
	var buffer = make([]uint8, rtl.DefaultBufLength)
	for {
		nRead, err := u.dev.ReadSync(buffer, rtl.DefaultBufLength)
		if err != nil {
			// log.Printf("\tReadSync Failed - error: %s\n", err)
			break
		}
		// log.Printf("\tReadSync %d\n", nRead)
		if nRead > 0 {
			// buf := buffer[:nRead]
			fmt.Printf("\rnRead %d: readCnt: %d", nRead, readCnt)
			readCnt++
		}
	}
}

func (u *UAT) scan(fStart int, fEnd int, aggregator chan []byte) {
	defer u.wg.Done()
	bandwith := 500 * 1000
	step := bandwith / 2
	buffer := make([]uint8, rtl.DefaultBufLength)
	for freq := fStart; freq <= fEnd; freq += step {
		nRead, err := u.dev.ReadSync(buffer, rtl.DefaultBufLength)
		if err != nil {
			break
		}

		if nRead > 0 {
			aggregator <- buffer
		}

	}
}

// sdrConfig configures the device to 978 MHz UAT channel.
func (u *UAT) sdrConfig(indexID int) (err error) {
	if u.dev, err = rtl.Open(indexID); err != nil {
		log.Printf("\tUAT Open Failed...\n")
		return
	}
	log.Printf("\tGetTunerType: %s\n", u.dev.GetTunerType())

	//---------- Set Tuner Gain ----------
	err = u.dev.SetTunerGainMode(true)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerGainMode Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetTunerGainMode Successful\n")

	tgain := 0
	gains, err := u.dev.GetTunerGains()
	if err != nil {
		log.Printf("\tGetTunerGains Failed - error: %s\n", err)
	} else if len(gains) > 0 {
		log.Printf("Gains: %v", gains)
		tgain = int(gains[0])
	}

	err = u.dev.SetTunerGain(tgain)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerGain Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetTunerGain Successful. Gain: %d\n", tgain)

	//---------- Get/Set Sample Rate ----------
	samplerate := 2048000 // 2000000 // 2083334
	err = u.dev.SetSampleRate(samplerate)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetSampleRate Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetSampleRate - rate: %d\n", samplerate)
	log.Printf("\tGetSampleRate: %d\n", u.dev.GetSampleRate())

	//---------- Get/Set Xtal Freq ----------
	rtlFreq, tunerFreq, err := u.dev.GetXtalFreq()
	if err != nil {
		u.dev.Close()
		log.Printf("\tGetXtalFreq Failed - error: %s\n", err)
		return
	}
	log.Printf("\tGetXtalFreq - Rtl: %d, Tuner: %d\n", rtlFreq, tunerFreq)

	newRTLFreq := 28800000
	newTunerFreq := 28800000
	err = u.dev.SetXtalFreq(newRTLFreq, newTunerFreq)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetXtalFreq Failed - error: %s\n", err)
		return
	}
	log.Printf("\tSetXtalFreq - Center freq: %d, Tuner freq: %d\n",
		newRTLFreq, newTunerFreq)

	//---------- Get/Set Center Freq ----------
	err = u.dev.SetCenterFreq(978000000)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetCenterFreq 978MHz Failed, error: %s\n", err)
		return
	}
	log.Printf("\tSetCenterFreq 978MHz Successful\n")

	log.Printf("\tGetCenterFreq: %d\n", u.dev.GetCenterFreq())

	//---------- Set Bandwidth ----------
	bw := 1000000
	log.Printf("\tSetting Bandwidth: %d\n", bw)
	if err = u.dev.SetTunerBw(bw); err != nil {
		u.dev.Close()
		log.Printf("\tSetTunerBw %d Failed, error: %s\n", bw, err)
		return
	}
	log.Printf("\tSetTunerBw %d Successful\n", bw)

	if err = u.dev.ResetBuffer(); err != nil {
		u.dev.Close()
		log.Printf("\tResetBuffer Failed - error: %s\n", err)
		return
	}
	log.Printf("\tResetBuffer Successful\n")

	//---------- Get/Set Freq Correction ----------
	freqCorr := u.dev.GetFreqCorrection()
	log.Printf("\tGetFreqCorrection: %d\n", freqCorr)
	err = u.dev.SetFreqCorrection(freqCorr)
	if err != nil {
		u.dev.Close()
		log.Printf("\tSetFreqCorrection %d Failed, error: %s\n", freqCorr, err)
		return
	}
	log.Printf("\tSetFreqCorrection %d Successful\n", freqCorr)

	return
}

