package s3plugin_test

import (
	"flag"
	"testing"

	"github.com/greenplum-db/gpbackup-s3-plugin/s3plugin"
	"github.com/urfave/cli"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "s3_plugin tests")
}

var _ = Describe("s3_plugin tests", func() {
	var pluginConfig *s3plugin.PluginConfig
	var opts *s3plugin.PluginOptions
	BeforeEach(func() {
		pluginConfig = &s3plugin.PluginConfig{
			ExecutablePath: "/tmp/location",
			Options: s3plugin.PluginOptions{
				AwsAccessKeyId:               "12345",
				AwsSecretAccessKey:           "6789",
				BackupMaxConcurrentRequests:  "5",
				BackupMultipartChunksize:     "7MB",
				Bucket:                       "bucket_name",
				Endpoint:                     "endpoint_name",
				Folder:                       "folder_name",
				Region:                       "region_name",
				RestoreMaxConcurrentRequests: "5",
				RestoreMultipartChunksize:    "7MB",
			},
		}
		opts = &pluginConfig.Options
	})
	Describe("GetS3Path", func() {
		It("it combines the folder directory with a path that results from removing all but the last 3 directories of the file path parameter", func() {
			folder := "s3/Dir"
			path := "/a/b/c/tmp/datadir/gpseg-1/backups/20180101/20180101082233/backup_file"
			newPath := s3plugin.GetS3Path(folder, path)
			expectedPath := "s3/Dir/backups/20180101/20180101082233/backup_file"
			Expect(newPath).To(Equal(expectedPath))
		})
	})
	Describe("ShouldEnableEncryption", func() {
		It("returns true when no encryption in config", func() {
			result := s3plugin.ShouldEnableEncryption("")
			Expect(result).To(BeTrue())
		})
		It("returns true when encryption set to 'on' in config", func() {
			result := s3plugin.ShouldEnableEncryption("on")
			Expect(result).To(BeTrue())
		})
		It("returns false when encryption set to 'off' in config", func() {
			result := s3plugin.ShouldEnableEncryption("off")
			Expect(result).To(BeFalse())
		})
		It("returns true when encryption set to anything else in config", func() {
			result := s3plugin.ShouldEnableEncryption("random_test")
			Expect(result).To(BeTrue())
		})
	})
	Describe("ValidateConfig", func() {
		It("succeeds when all fields in config filled", func() {
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(BeNil())
		})
		It("succeeds when all fields except endpoint filled in config", func() {
			opts.Endpoint = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(BeNil())
		})
		It("succeeds when all fields except region filled in config", func() {
			opts.Region = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(BeNil())
		})
		It("succeeds when all fields except aws_access_key_id and aws_secret_access_key in config", func() {
			opts.AwsAccessKeyId = ""
			opts.AwsSecretAccessKey = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(BeNil())
		})
		It("sets region to unused when endpoint is used instead of region", func() {
			opts.Region = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).ToNot(HaveOccurred())
			Expect(pluginConfig.Options.Region).To(Equal("unused"))
		})
		It("returns error when neither region nor endpoint in config", func() {
			opts.Region = ""
			opts.Endpoint = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(HaveOccurred())
		})
		It("returns error when no aws_access_key_id in config", func() {
			opts.AwsAccessKeyId = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(HaveOccurred())
		})
		It("returns error when no aws_secret_access_key in config", func() {
			opts.AwsSecretAccessKey = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(HaveOccurred())
		})
		It("returns error when no bucket in config", func() {
			opts.Bucket = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(HaveOccurred())
		})
		It("returns error when no folder in config", func() {
			opts.Folder = ""
			err := s3plugin.ValidateConfig(pluginConfig)
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Optional config params", func() {
		It("correctly parses upload params from config", func() {
			chunkSize, err := s3plugin.GetUploadChunkSize(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(chunkSize).To(Equal(int64(7 * 1024 * 1024)))

			concurrency, err := s3plugin.GetUploadConcurrency(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(concurrency).To(Equal(5))
		})
		It("uses default values if upload params are not specified", func() {
			opts.BackupMultipartChunksize = ""
			opts.BackupMaxConcurrentRequests = ""

			chunkSize, err := s3plugin.GetUploadChunkSize(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(chunkSize).To(Equal(int64(500 * 1024 * 1024)))

			concurrency, err := s3plugin.GetUploadConcurrency(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(concurrency).To(Equal(6))
		})
		It("correctly parses download params from config", func() {
			chunkSize, err := s3plugin.GetDownloadChunkSize(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(chunkSize).To(Equal(int64(7 * 1024 * 1024)))

			concurrency, err := s3plugin.GetDownloadConcurrency(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(concurrency).To(Equal(5))
		})
		It("uses default values if download params are not specified", func() {
			opts.RestoreMultipartChunksize = ""
			opts.RestoreMaxConcurrentRequests = ""

			chunkSize, err := s3plugin.GetDownloadChunkSize(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(chunkSize).To(Equal(int64(500 * 1024 * 1024)))

			concurrency, err := s3plugin.GetDownloadConcurrency(pluginConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(concurrency).To(Equal(6))
		})
	})
	Describe("Delete", func() {
		var flags *flag.FlagSet

		BeforeEach(func() {
			flags = flag.NewFlagSet("testing flagset", flag.PanicOnError)
		})
		It("returns error when timestamp is not provided", func() {
			err := flags.Parse([]string{"myconfigfilepath"})
			Expect(err).ToNot(HaveOccurred())
			context := cli.NewContext(nil, flags, nil)

			err = s3plugin.DeleteBackup(context)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("delete requires a <timestamp>"))
		})
		It("returns error when timestamp does not parse", func() {
			err := flags.Parse([]string{"myconfigfilepath", "badformat"})
			Expect(err).ToNot(HaveOccurred())
			context := cli.NewContext(nil, flags, nil)

			err = s3plugin.DeleteBackup(context)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("delete requires a <timestamp> with format YYYYMMDDHHMMSS, but received: badformat"))
		})
	})
})
