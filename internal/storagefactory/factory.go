package storagefactory

import (
    "github.com/langgenius/dify-cloud-kit/oss"
    "github.com/langgenius/dify-cloud-kit/oss/factory"
    "github.com/langgenius/dify-plugin-daemon/internal/types/app"
)

// argsFromConfig maps app.Config to oss.OSSArgs.
func argsFromConfig(cfg *app.Config) oss.OSSArgs {
    return oss.OSSArgs{
        Local: &oss.Local{Path: cfg.PluginStorageLocalRoot},
        S3: &oss.S3{
            UseAws:       cfg.S3UseAWS,
            Endpoint:     cfg.S3Endpoint,
            UsePathStyle: cfg.S3UsePathStyle,
            AccessKey:    cfg.AWSAccessKey,
            SecretKey:    cfg.AWSSecretKey,
            Bucket:       cfg.PluginStorageOSSBucket,
            Region:       cfg.AWSRegion,
            UseIamRole:   cfg.S3UseAwsManagedIam,
        },
        TencentCOS: &oss.TencentCOS{
            Region:    cfg.TencentCOSRegion,
            SecretID:  cfg.TencentCOSSecretId,
            SecretKey: cfg.TencentCOSSecretKey,
            Bucket:    cfg.PluginStorageOSSBucket,
        },
        AzureBlob: &oss.AzureBlob{
            ConnectionString: cfg.AzureBlobStorageConnectionString,
            ContainerName:    cfg.AzureBlobStorageContainerName,
        },
        GoogleCloudStorage: &oss.GoogleCloudStorage{
            Bucket:         cfg.PluginStorageOSSBucket,
            CredentialsB64: cfg.GoogleCloudStorageCredentialsB64,
        },
        AliyunOSS: &oss.AliyunOSS{
            Region:      cfg.AliyunOSSRegion,
            Endpoint:    cfg.AliyunOSSEndpoint,
            AccessKey:   cfg.AliyunOSSAccessKeyID,
            SecretKey:   cfg.AliyunOSSAccessKeySecret,
            AuthVersion: cfg.AliyunOSSAuthVersion,
            Path:        cfg.AliyunOSSPath,
            Bucket:      cfg.PluginStorageOSSBucket,
        },
        HuaweiOBS: &oss.HuaweiOBS{
            AccessKey: cfg.HuaweiOBSAccessKey,
            SecretKey: cfg.HuaweiOBSSecretKey,
            Server:    cfg.HuaweiOBSServer,
            Bucket:    cfg.PluginStorageOSSBucket,
        },
        VolcengineTOS: &oss.VolcengineTOS{
            Region:    cfg.VolcengineTOSRegion,
            Endpoint:  cfg.VolcengineTOSEndpoint,
            AccessKey: cfg.VolcengineTOSAccessKey,
            SecretKey: cfg.VolcengineTOSSecretKey,
            Bucket:    cfg.PluginStorageOSSBucket,
        },
    }
}

// New constructs an oss.OSS for the given type using the provided config.
// Returns the instance and any error encountered.
func New(t string, cfg *app.Config) (oss.OSS, error) {
    s, err := factory.Load(t, argsFromConfig(cfg))
    if err != nil {
        return nil, err
    }
    return s, nil
}
