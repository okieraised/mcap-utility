package cmd

import (
	"fmt"
	"github.com/foxglove/mcap/go/mcap"
	"github.com/spf13/cobra"
	"io"
	"mcap-cli/internal/constants"
	"mcap-cli/internal/logging"
	"mcap-cli/internal/utils"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

var writerOpt = &mcap.WriterOptions{
	IncludeCRC:               true,
	Chunked:                  true,
	Compression:              mcap.CompressionZSTD,
	CompressionLevel:         mcap.CompressionLevelDefault,
	SkipMessageIndexing:      false,
	SkipStatistics:           false,
	SkipRepeatedSchemas:      false,
	SkipRepeatedChannelInfos: false,
	SkipAttachmentIndex:      false,
	SkipMetadataIndex:        false,
	SkipChunkIndex:           false,
	SkipSummaryOffsets:       false,
	OverrideLibrary:          false,
	SkipMagic:                false,
}

var (
	input            string
	output           string
	rename           map[string]string
	trimStart        string
	trimEnd          string
	shiftLog         string
	shiftPublish     string
	compression      string
	topics           []string
	deletes          []string
	usePubTime       bool
	compressionLevel int
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: fmt.Sprintf("Edit the contents of (%s) file or directory", constants.MCAPFIleExtension),
	Long:  fmt.Sprintf("Edit the contents of (%s) file or directory", constants.MCAPFIleExtension),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		compressionMapper := map[string]struct{}{
			"lz4":  {},
			"zstd": {},
		}

		if compression != "" {
			compression = strings.TrimSpace(strings.ToLower(compression))
			if _, ok := compressionMapper[compression]; !ok {
				return fmt.Errorf("invalid compression: %s", compression)
			}
			writerOpt.Compression = mcap.CompressionFormat(compression)
		}

		compressionLevelMapper := map[int]mcap.CompressionLevel{
			0: mcap.CompressionLevelDefault,
			1: mcap.CompressionLevelFastest,
			2: mcap.CompressionLevelBetter,
			3: mcap.CompressionLevelBest,
		}

		if compressionLevel != 0 {
			if _, ok := compressionLevelMapper[compressionLevel]; !ok {
				return fmt.Errorf("invalid compression level: %d", compressionLevel)
			}
			writerOpt.CompressionLevel = compressionLevelMapper[compressionLevel]
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	editCmd.
		Flags().
		StringVarP(
			&input,
			"input",
			"i",
			"",
			fmt.Sprintf(
				"Input (%s) files or directory contains (%s) file(s)",
				constants.MCAPFIleExtension,
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&output,
			"output",
			"o",
			"",
			fmt.Sprintf(
				"Output directory to save processed (%s) files",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.Flags().
		StringToStringVarP(
			&rename,
			"rename",
			"r",
			nil,
			fmt.Sprintf(
				"Topic mappings to rename inside (%s) file (/old_topic_1=/new_topic_1,old_topic_2=/new_topic_2)",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&trimStart,
			"trim-start",
			"s",
			"",
			fmt.Sprintf(
				"Timestamp at which to start trimming (%s) file (prefer RFC3339 or unixnano format, e.g. 2006-01-02T15:04:05+07:00)",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&trimEnd,
			"trim-end",
			"e",
			"",
			fmt.Sprintf(
				"Timestamp at which to end trimming (%s) file (prefer RFC3339 or unixnano format, e.g. 2006-01-02T15:04:05+07:00))",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&shiftLog,
			"shift-log",
			"l",
			"",
			fmt.Sprintf(
				"Duration to shift message log time inside (%s) file (e.g. 100ms,10ns,10m30s,-1h)",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&shiftPublish,
			"shift-pub",
			"p",
			"",
			fmt.Sprintf(
				"Duration to shift message publish time inside (%s) file (e.g. 100ms,10ns,10m30s,-1h)",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringSliceVarP(
			&topics,
			"topics",
			"t",
			nil,
			fmt.Sprintf(
				"List of topics to perform action on (applied to shift-log and shift-pub), if unspecified, shift will be applied to all topics",
			),
		)

	editCmd.
		Flags().
		StringSliceVarP(
			&deletes,
			"delete",
			"d",
			nil,
			fmt.Sprintf(
				"List of topics to remove from (%s) files",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		BoolVarP(
			&usePubTime,
			"pub-time",
			"b",
			false,
			fmt.Sprintf(
				"Change the time of the ROS Timestamp to the time the message was published in (%s) files",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		StringVarP(
			&compression,
			"compression",
			"c",
			"",
			fmt.Sprintf(
				"Compression algorithm used to write (%s) files (zstd or lz4)",
				constants.MCAPFIleExtension,
			),
		)

	editCmd.
		Flags().
		IntVarP(
			&compressionLevel,
			"compression-level",
			"n",
			0,
			fmt.Sprintf(
				"Compression level write (%s) files (0:default 1:fastest 2:better 3:best)",
				constants.MCAPFIleExtension,
			),
		)

	_ = editCmd.MarkFlagRequired("input")
	_ = editCmd.MarkFlagRequired("output")
}

func run() {
	if len(rename) == 0 && len(deletes) == 0 && trimStart == "" &&
		trimEnd == "" && len(topics) == 0 && shiftPublish == "" &&
		shiftLog == "" && !usePubTime && compression == "" {
		logging.GetLogger().Info("Nothing to do")
		os.Exit(0)
	}

	isDir, err := utils.IsPathDirectory(input)
	if err != nil {
		logging.GetLogger().Error(err.Error())
		os.Exit(1)
	}

	fileToProcess := make([]string, 0, 10)

	if !isDir {
		logging.GetLogger().Info("Input path is not a directory")
		if !strings.HasSuffix(input, constants.MCAPFIleExtension) {
			logging.GetLogger().Info(fmt.Sprintf("Input %s does not end with %s extension", input, constants.MCAPFIleExtension))
			os.Exit(1)
		}
		fileToProcess = append(fileToProcess, input)
	} else {
		logging.GetLogger().Info("Input path is a directory")
		mcapFiles, err := utils.ListMCAPFilesInDirectory(input)
		if err != nil {
			logging.GetLogger().Error(err.Error())
			os.Exit(1)
		}
		fileToProcess = append(fileToProcess, mcapFiles...)
	}
	logging.GetLogger().Info(fmt.Sprintf("Retrieved %d mcap files to process", len(fileToProcess)))

	if len(fileToProcess) == 0 {
		logging.GetLogger().Info(fmt.Sprintf("No (%s) files to process", constants.MCAPFIleExtension))
		os.Exit(0)
	}

	isSameDir, err := utils.IsSameDirectory(input, output)
	if err != nil {
		logging.GetLogger().Error(fmt.Sprintf("Unable to determine current directory: %s", err))
		os.Exit(1)
	}

	if isSameDir {
		logging.GetLogger().Info("Cannot use input directory as output directory")
		os.Exit(1)
	}

	exist := utils.IsDirExists(output)
	if !exist {
		logging.GetLogger().Info("Output directory does not exist")
		err := utils.CreateDir(output)
		if err != nil {
			logging.GetLogger().Error(err.Error())
			os.Exit(1)
		}
		logging.GetLogger().Info("Output directory created")
	}

	err = process(fileToProcess)
	if err != nil {
		logging.GetLogger().Error(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func process(toProcess []string) error {
	dataCh := make(chan string, len(toProcess))
	stopCh := make(chan struct{})
	errorCh := make(chan error, 1)

	var wg sync.WaitGroup

	go func(data []string, dataCh chan<- string, stopCh <-chan struct{}) {
		defer close(dataCh)
		for _, item := range data {
			select {
			case <-stopCh:
				return
			default:
				dataCh <- item
			}
		}
	}(toProcess, dataCh, stopCh)

	numConsumers := runtime.NumCPU()*2 + 1
	for i := 1; i <= numConsumers; i++ {
		wg.Add(1)

		go func(idx int, dataCh <-chan string, wg *sync.WaitGroup, stopCh <-chan struct{}, errorCh chan<- error) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				case fPath, ok := <-dataCh:
					if !ok {
						return
					}
					logging.GetLogger().Info(fmt.Sprintf("Processing %s", fPath))
					err := conversion(fPath)
					if err != nil {
						errorCh <- err
					}
				}
			}
		}(i, dataCh, &wg, stopCh, errorCh)
	}

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	select {
	case err := <-errorCh:
		if err != nil {
			close(stopCh)
			return err
		}
	case <-stopCh:
		logging.GetLogger().Info("Stop signal received. Stopping.")
	}

	// Wait for consumers to clean up
	wg.Wait()
	logging.GetLogger().Info("All consumers have completed, exiting...")
	return nil
}

func conversion(filePath string) error {
	inFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(inFile *os.File) {
		cErr := inFile.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}(inFile)

	reader, err := mcap.NewReader(inFile)
	if err != nil {
		return fmt.Errorf("failed to create new reader for %s: %s", filePath, err)
	}

	baseFileName := filepath.Base(filePath)
	outputPath := filepath.Join(output, baseFileName)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %s", outputPath, err)
	}
	defer func(outFile *os.File) {
		cErr := inFile.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}(outFile)

	writer, err := mcap.NewWriter(outFile, writerOpt)
	if err != nil {
		return fmt.Errorf("failed to create new writer: %s", err)
	}

	msgs, err := reader.Messages()
	if err != nil {
		return fmt.Errorf("failed to read messages: %s", err)
	}

	err = writer.WriteHeader(reader.Header())
	if err != nil {
		return fmt.Errorf("failed to write header: %s", err)
	}

	schemaWritten := map[uint16]bool{}
	channelWritten := map[uint16]bool{}
	channelMap := map[uint16]uint16{}

	mcapInfo, err := reader.Info()
	if err != nil {
		return fmt.Errorf("failed to read mcap info: %s", err)
	}

	msgLogStart := mcapInfo.Statistics.MessageStartTime
	msgLogEnd := mcapInfo.Statistics.MessageEndTime

	channelIDs := make([]uint16, 0, len(mcapInfo.Channels))
	for _, channel := range mcapInfo.Channels {
		channelIDs = append(channelIDs, channel.ID)
	}

	id, err := utils.RandomUint16NotIn(channelIDs)
	if err != nil {
		return err
	}

	// Perform trimming if specified
	var trimStartTime, trimEndTime int64

	if trimStart != "" {
		trimStartTime, err = utils.TryParseTimestamp(trimStart)
		if err != nil {
			return fmt.Errorf("invalid trim start time: %s", trimStart)
		}
		if trimStartTime < int64(msgLogStart) {
			return fmt.Errorf("trim start time [%d] is before message start time [%d]", trimStartTime, msgLogStart)
		}

		if trimStartTime >= int64(msgLogEnd) {
			return fmt.Errorf("trim start time [%d] is after message end time [%d]", trimStartTime, msgLogEnd)
		}
	}

	if trimEnd != "" {
		trimEndTime, err = utils.TryParseTimestamp(trimEnd)
		if err != nil {
			return fmt.Errorf("invalid trim end time: %s", trimEnd)
		}
		if trimEndTime > int64(msgLogEnd) {
			logging.GetLogger().Warn(fmt.Sprintf("trim end time [%d] is after message end time [%d]", trimStartTime, msgLogEnd))
		}
	} else {
		trimEndTime = int64(msgLogEnd)
	}

	for {
		schema, channel, msg, err := msgs.NextInto(&mcap.Message{})
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to iterate messages: %s", err)
		}

		if trimStartTime != 0 {
			logTime := int64(msg.LogTime)

			if logTime < trimStartTime {
				continue
			}

			if logTime > trimEndTime {
				break
			}
		}

		// Perform remove if specified
		if len(deletes) > 0 {
			if slices.Contains(deletes, channel.Topic) {
				continue
			}
		}

		if !schemaWritten[channel.SchemaID] {
			if err := writer.WriteSchema(schema); err != nil {
				return err
			}
			schemaWritten[channel.SchemaID] = true
		}
		newChannelID := channel.ID

		// Perform rename if specified
		if len(rename) != 0 {
			for oldTopic, newTopic := range rename {
				if !strings.HasPrefix(newTopic, "/") {
					return fmt.Errorf("invalid new topic name: %s", newTopic)
				}

				if channel.Topic == oldTopic {
					newChannelID = id
					newChannel := &mcap.Channel{
						ID:              newChannelID,
						SchemaID:        channel.SchemaID,
						Topic:           newTopic,
						MessageEncoding: channel.MessageEncoding,
						Metadata:        channel.Metadata,
					}
					if err := writer.WriteChannel(newChannel); err != nil {
						return err
					}
					channelMap[channel.ID] = newChannelID
				} else if _, exists := channelMap[channel.ID]; !exists {
					if err := writer.WriteChannel(channel); err != nil {
						return err
					}
					channelMap[channel.ID] = channel.ID
				}
			}
			msg.ChannelID = channelMap[channel.ID]
		} else {
			if !schemaWritten[channel.SchemaID] {
				if err := writer.WriteSchema(schema); err != nil {
					return fmt.Errorf("write schema: %w", err)
				}
				schemaWritten[channel.SchemaID] = true
			}

			if !channelWritten[channel.ID] {
				if err := writer.WriteChannel(channel); err != nil {
					return fmt.Errorf("write channel: %w", err)
				}
				channelWritten[channel.ID] = true
			}
		}

		// Shift log time if applicable
		if shiftLog != "" {
			duration, err := time.ParseDuration(shiftLog)
			if err != nil {
				return fmt.Errorf("invalid shift log time %s: %s", shiftLog, err)
			}

			if len(topics) > 0 {
				if slices.Contains(topics, channel.Topic) {
					msg.LogTime = uint64(int64(msg.LogTime) + int64(duration))
				}
			} else {
				msg.LogTime = uint64(int64(msg.LogTime) + int64(duration))
			}
		}

		// shift publish time if application
		if shiftPublish != "" {
			duration, err := time.ParseDuration(shiftPublish)
			if err != nil {
				return fmt.Errorf("invalid shift publish time %s: %s", shiftPublish, err)
			}

			if len(topics) > 0 {
				if slices.Contains(topics, channel.Topic) {
					msg.PublishTime = uint64(int64(msg.PublishTime) + int64(duration))
				}
			} else {
				msg.PublishTime = uint64(int64(msg.PublishTime) + int64(duration))
			}
		}

		if err := writer.WriteMessage(msg); err != nil {
			return err
		}
	}

	return writer.Close()
}
