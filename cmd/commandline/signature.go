package main

import (
	"os"
	"strings"

	"github.com/langgenius/dify-plugin-daemon/cmd/commandline/signature"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/spf13/cobra"
)

var (
	signatureGenerateCommand = &cobra.Command{
		Use:   "generate",
		Short: "Generate a key pair",
		Long:  "Generate a key pair",
		Args:  cobra.ExactArgs(0),
		Run: func(c *cobra.Command, args []string) {
			keyPairName := c.Flag("filename").Value.String()
			if keyPairName == "" {
				keyPairName = "dify_plugin_signing_key"
			}
			err := signature.GenerateKeyPair(keyPairName)
			if err != nil {
				os.Exit(1)
			}
		},
	}

	signatureSignCommand = &cobra.Command{
		Use:   "sign [difypkg_path]",
		Short: "Sign a difypkg file",
		Long:  "Sign a difypkg file with the specified private key",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			difypkgPath := args[0]
			privateKeyPath := c.Flag("private_key").Value.String()
			authorizedCategory := c.Flag("authorized_category").Value.String()
			if authorizedCategory != "" {
				if !strings.EqualFold(authorizedCategory, string(decoder.AUTHORIZED_CATEGORY_LANGGENIUS)) &&
					!strings.EqualFold(authorizedCategory, string(decoder.AUTHORIZED_CATEGORY_PARTNER)) &&
					!strings.EqualFold(authorizedCategory, string(decoder.AUTHORIZED_CATEGORY_COMMUNITY)) {
					log.Error("invalid authorized category: %s", authorizedCategory)
					os.Exit(1)
				}
			}

			err := signature.Sign(difypkgPath, privateKeyPath, &decoder.Verification{
				AuthorizedCategory: decoder.AuthorizedCategory(authorizedCategory),
			})
			if err != nil {
				os.Exit(1)
			}
		},
	}

	signatureVerifyCommand = &cobra.Command{
		Use:   "verify [difypkg_path]",
		Short: "Verify a difypkg file",
		Long:  "Verify a difypkg file with the specified public key. If no public key is provided, the official public key will be used",
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			difypkgPath := args[0]
			publicKeyPath := c.Flag("public_key").Value.String()
			err := signature.Verify(difypkgPath, publicKeyPath)
			if err != nil {
				os.Exit(1)
			}
		},
	}
)

func init() {
	signatureCommand.AddCommand(signatureGenerateCommand)
	signatureCommand.AddCommand(signatureSignCommand)
	signatureCommand.AddCommand(signatureVerifyCommand)

	signatureGenerateCommand.Flags().StringP("filename", "f", "", "filename of the key pair")

	signatureSignCommand.Flags().StringP("private_key", "p", "", "private key file")
	signatureSignCommand.MarkFlagRequired("private_key")

	signatureSignCommand.Flags().StringP(
		"authorized_category",
		"c",
		string(decoder.AUTHORIZED_CATEGORY_LANGGENIUS),
		"authorized category",
	)

	signatureVerifyCommand.Flags().StringP("public_key", "p", "", "public key file")
}
