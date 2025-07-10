package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// ListBuckets returns a list of all buckets with detailed statistics
func (s *MinIOService) ListBuckets(ctx context.Context, username, password string) ([]BucketInfo, error) {
	log.Printf("[DEBUG] MinIO service ListBuckets called for user '%s'", username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListBuckets: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListBuckets API")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListBuckets API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] MinIO ListBuckets API returned %d buckets", len(buckets))

	var bucketInfos []BucketInfo
	for _, bucket := range buckets {
		info := BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate.Format("2006-01-02 15:04:05"),
			Size:         0, // Will be calculated below
			ObjectCount:  0, // Will be calculated below
		}

		// Get bucket statistics (size and object count)
		log.Printf("[DEBUG] Getting statistics for bucket '%s'", bucket.Name)
		size, objectCount := s.getBucketStats(ctx, client, bucket.Name)
		info.Size = size
		info.ObjectCount = objectCount

		bucketInfos = append(bucketInfos, info)
		log.Printf("[DEBUG] Bucket: %s (created: %s, size: %d bytes, objects: %d)",
			bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"), size, objectCount)
	}

	log.Printf("[DEBUG] Returning %d bucket infos", len(bucketInfos))
	return bucketInfos, nil
}

// ListBucketsQuick provides a faster bucket listing without size/count calculation
func (s *MinIOService) ListBucketsQuick(ctx context.Context, username, password string) ([]BucketInfo, error) {
	log.Printf("[DEBUG] MinIO service ListBucketsQuick called for user '%s'", username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in ListBucketsQuick: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Calling MinIO ListBuckets API (quick mode)")
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Printf("[DEBUG] MinIO ListBuckets API failed: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] MinIO ListBuckets API returned %d buckets", len(buckets))

	var bucketInfos []BucketInfo
	for _, bucket := range buckets {
		info := BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate.Format("2006-01-02 15:04:05"),
			Size:         -1, // -1 indicates not calculated
			ObjectCount:  -1, // -1 indicates not calculated
		}
		bucketInfos = append(bucketInfos, info)
		log.Printf("[DEBUG] Bucket: %s (created: %s, stats: not calculated)",
			bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
	}

	log.Printf("[DEBUG] Returning %d bucket infos (quick mode)", len(bucketInfos))
	return bucketInfos, nil
}

// getBucketStats calculates the total size and object count for a bucket
func (s *MinIOService) getBucketStats(ctx context.Context, client *minio.Client, bucketName string) (int64, int64) {
	var totalSize int64
	var objectCount int64

	// Create a timeout context for bucket stats calculation (30 seconds max)
	statsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("[DEBUG] Starting stats calculation for bucket '%s' (max 30s timeout)", bucketName)

	// List all objects in the bucket to calculate size and count
	objectCh := client.ListObjects(statsCtx, bucketName, minio.ListObjectsOptions{Recursive: true})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("[DEBUG] Error listing object in bucket '%s': %v", bucketName, object.Err)
			// Check if it's a timeout or context cancellation
			if statsCtx.Err() != nil {
				log.Printf("[DEBUG] Stats calculation timed out for bucket '%s'", bucketName)
				return -1, -1 // Return -1 to indicate timeout/error
			}
			continue
		}
		totalSize += object.Size
		objectCount++

		// Check for timeout periodically
		select {
		case <-statsCtx.Done():
			log.Printf("[DEBUG] Stats calculation timed out for bucket '%s' after %d objects", bucketName, objectCount)
			return -1, -1
		default:
			// Continue processing
		}
	}

	log.Printf("[DEBUG] Stats calculation completed for bucket '%s': %d bytes, %d objects",
		bucketName, totalSize, objectCount)
	return totalSize, objectCount
}

// CreateBucket creates a new bucket
func (s *MinIOService) CreateBucket(ctx context.Context, bucketName, username, password string) error {
	client, _, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

// DeleteBucket deletes an existing bucket
func (s *MinIOService) DeleteBucket(ctx context.Context, bucketName, username, password string) error {
	client, _, err := s.CreateClients(username, password)
	if err != nil {
		return err
	}
	return client.RemoveBucket(ctx, bucketName)
}

// GetBucketPolicy returns the bucket policy as a JSON string
func (s *MinIOService) GetBucketPolicy(ctx context.Context, bucketName, username, password string) (string, error) {
	log.Printf("[DEBUG] MinIO service GetBucketPolicy called for bucket '%s' by user '%s'", bucketName, username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetBucketPolicy: %v", err)
		return "", err
	}

	log.Printf("[DEBUG] Calling MinIO GetBucketPolicy API for bucket '%s'", bucketName)
	policy, err := client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		log.Printf("[DEBUG] MinIO GetBucketPolicy API response for bucket '%s': %v", bucketName, err)

		// Check if this is a "no policy" error, which is normal
		errStr := err.Error()
		if strings.Contains(errStr, "policy does not exist") ||
			strings.Contains(errStr, "NoSuchBucketPolicy") ||
			strings.Contains(errStr, "The bucket policy does not exist") {
			log.Printf("[DEBUG] No policy exists for bucket '%s' (this is normal)", bucketName)
			return "", nil // Return empty string, not an error
		}

		// For other errors, return the actual error
		log.Printf("[DEBUG] GetBucketPolicy failed for bucket '%s' with error: %v", bucketName, err)
		return "", err
	}

	log.Printf("[DEBUG] GetBucketPolicy successful for bucket '%s', policy length: %d", bucketName, len(policy))
	return policy, nil
}

// SetBucketPolicy sets the bucket policy
func (s *MinIOService) SetBucketPolicy(ctx context.Context, bucketName, policy, username, password string) error {
	log.Printf("[DEBUG] MinIO service SetBucketPolicy called for bucket '%s' by user '%s', policy length: %d", bucketName, username, len(policy))

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in SetBucketPolicy: %v", err)
		return err
	}

	log.Printf("[DEBUG] Calling MinIO SetBucketPolicy API for bucket '%s'", bucketName)
	err = client.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Printf("[DEBUG] MinIO SetBucketPolicy API failed for bucket '%s': %v", bucketName, err)
		return err
	}

	log.Printf("[DEBUG] SetBucketPolicy successful for bucket '%s'", bucketName)
	return nil
}

// GetBucketStatsQuick returns bucket statistics with a shorter timeout for dashboard use
func (s *MinIOService) GetBucketStatsQuick(ctx context.Context, username, password, bucketName string) (int64, int64) {
	log.Printf("[DEBUG] GetBucketStatsQuick called for bucket '%s' by user '%s'", bucketName, username)

	client, _, err := s.CreateClients(username, password)
	if err != nil {
		log.Printf("[DEBUG] Failed to create clients in GetBucketStatsQuick: %v", err)
		return -1, -1
	}

	// Use shorter timeout for dashboard (5 seconds max)
	return s.getBucketStatsWithTimeout(ctx, client, bucketName, 5*time.Second)
}

// getBucketStatsWithTimeout calculates bucket stats with configurable timeout
func (s *MinIOService) getBucketStatsWithTimeout(ctx context.Context, client *minio.Client, bucketName string, timeout time.Duration) (int64, int64) {
	var totalSize int64
	var objectCount int64

	// Create a timeout context for bucket stats calculation
	statsCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Printf("[DEBUG] Starting quick stats calculation for bucket '%s' (max %.0fs timeout)", bucketName, timeout.Seconds())

	// List all objects in the bucket to calculate size and count
	objectCh := client.ListObjects(statsCtx, bucketName, minio.ListObjectsOptions{Recursive: true})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("[DEBUG] Error listing object in bucket '%s': %v", bucketName, object.Err)
			// Check if it's a timeout or context cancellation
			if statsCtx.Err() != nil {
				log.Printf("[DEBUG] Quick stats calculation timed out for bucket '%s'", bucketName)
				return -1, -1 // Return -1 to indicate timeout/error
			}
			continue
		}
		totalSize += object.Size
		objectCount++

		// Check for timeout periodically
		select {
		case <-statsCtx.Done():
			log.Printf("[DEBUG] Quick stats calculation timed out for bucket '%s' after %d objects", bucketName, objectCount)
			return -1, -1
		default:
			// Continue processing
		}
	}

	log.Printf("[DEBUG] Quick stats calculation completed for bucket '%s': %d bytes, %d objects",
		bucketName, totalSize, objectCount)
	return totalSize, objectCount
}
