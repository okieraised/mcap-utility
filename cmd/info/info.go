package info

import (
	"fmt"
	"github.com/foxglove/mcap/go/mcap"
	"github.com/spf13/cobra"
	"mcap-utility/internal/constants"
	"mcap-utility/internal/logging"
	"os"
)

type topicInfo struct {
	topicID        uint16
	name           string
	messageCount   uint64
	topicEncoding  string
	schemaID       uint16
	schemaName     string
	schemaEncoding string
}

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: fmt.Sprintf("List the detail of a (%s) file", constants.MCAPFIleExtension),
	Long:  fmt.Sprintf("List the detail of a (%s) file", constants.MCAPFIleExtension),
	Run: func(cmd *cobra.Command, args []string) {
		info()
	},
}

var (
	topicInfoMapper = map[string]*topicInfo{}
	file            string
)

func init() {
	InfoCmd.
		Flags().
		StringVarP(
			&file,
			"file",
			"f",
			"",
			fmt.Sprintf(
				"Input (%s) file to retrieve information.",
				constants.MCAPFIleExtension,
			),
		)

	_ = InfoCmd.MarkFlagRequired("file")
}

func info() {
	mcapFile, err := os.Open(file)
	if err != nil {
		logging.GetLogger().Error(err.Error())
		os.Exit(1)
	}
	defer func(mcapFile *os.File) {
		cErr := mcapFile.Close()
		if cErr != nil && err == nil {
			err = cErr
		}
	}(mcapFile)

	reader, err := mcap.NewReader(mcapFile)
	if err != nil {
		logging.GetLogger().Error(err.Error())
		os.Exit(1)
	}
	defer reader.Close()

	info, err := reader.Info()
	if err != nil {
		logging.GetLogger().Error(err.Error())
		os.Exit(1)
	}

	for _, channel := range info.Channels {
		topicInfoMapper[channel.Topic] = &topicInfo{
			topicID:        channel.ID,
			name:           channel.Topic,
			messageCount:   info.ChannelCounts()[channel.Topic],
			topicEncoding:  channel.MessageEncoding,
			schemaID:       channel.SchemaID,
			schemaName:     info.Schemas[channel.SchemaID].Name,
			schemaEncoding: info.Schemas[channel.SchemaID].Encoding,
		}
	}

	fmt.Println(fmt.Sprintf("Header Library:         %s", reader.Header().Library))
	fmt.Println(fmt.Sprintf("Header Profile:         %s", reader.Header().Profile))
	fmt.Println(fmt.Sprintf("Summary Start:          %d", info.Footer.SummaryStart))
	fmt.Println(fmt.Sprintf("Summary Offset Start:   %d", info.Footer.SummaryOffsetStart))
	fmt.Println(fmt.Sprintf("Summary CRC:            %d", info.Footer.SummaryCRC))
	fmt.Println(fmt.Sprintf("Schema Count:           %d", info.Statistics.SchemaCount))
	fmt.Println(fmt.Sprintf("Chunk Count:            %d", info.Statistics.ChunkCount))
	fmt.Println(fmt.Sprintf("Metadata Count:         %d", info.Statistics.MetadataCount))
	fmt.Println(fmt.Sprintf("Attachment Count:       %d", info.Statistics.AttachmentCount))
	fmt.Println(fmt.Sprintf("Message Start Time:     %d", info.Statistics.MessageStartTime))
	fmt.Println(fmt.Sprintf("Message End Time:       %d", info.Statistics.MessageEndTime))
	fmt.Println(fmt.Sprintf("Message Count:          %d", info.Statistics.MessageCount))
	fmt.Println(fmt.Sprintf("Metadata Index Count:   %d", len(info.MetadataIndexes)))
	fmt.Println(fmt.Sprintf("Attachment Index Count: %d", len(info.AttachmentIndexes)))
	fmt.Println(fmt.Sprintf("Chunk Index Count:      %d", len(info.ChunkIndexes)))
	fmt.Println(fmt.Sprintf("Topic Count:            %d", info.Statistics.ChannelCount))

	for _, metadata := range info.MetadataIndexes {
		fmt.Println(fmt.Sprintf("Metadata Name: %s | Offset: %d | Length: %d", metadata.Name, metadata.Offset, metadata.Length))
	}

	for _, v := range topicInfoMapper {
		fmt.Println(fmt.Sprintf("Topic ID: %d | Topic: %s | Message Count: %d | Topic Encoding: %s | "+
			"Schema ID: %d | Schema Name: %s | Schema Encoding: %s",
			v.topicID, v.name, v.messageCount, v.topicEncoding, v.schemaID, v.schemaName, v.schemaEncoding))
	}

	os.Exit(0)
}
