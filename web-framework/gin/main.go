package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
)

func main() {
	router := gin.Default()

	folderPath := "./files"

	router.GET("/files", func(c *gin.Context) {
		files, err := os.ReadDir(folderPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Extract and return only filenames, excluding paths
		var filenames []string
		for _, file := range files {
			filenames = append(filenames, file.Name())
		}

		c.JSON(http.StatusOK, gin.H{"filenames": filenames})
	})

	router.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		// Validate filename (optional)
		// Implement validation logic here to ensure safe filenames (e.g., alphanumeric and underscore)

		filePath := path.Join(folderPath, filename)

		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		defer file.Close()

		// Set content headers
		c.Header("Content-Type", "application/octet-stream") // adjust content type based on file type
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		// Stream the file content
		_, err = io.Copy(c.Writer, file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error downloading file"})
			return
		}
	})

	getIPAddress()

	// Start the server
	log.Println("Server started at :8080")
	log.Fatal(router.Run(":8080"))
}

func getIPAddress() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Iterate through the interfaces
	for _, iface := range interfaces {
		// Check if the interface is up and not a loopback or virtual interface
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagPointToPoint == 0 {
			// Get the addresses associated with the interface
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			// Iterate through the addresses
			for _, addr := range addrs {
				// Check if the address is an IPv4 address
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						fmt.Printf("Interface: %s, IPv4 Address: %s\n", iface.Name, ipnet.IP.String())
					}
				}
			}
		}
	}
}
