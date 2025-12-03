use std::io;

use insane_sdfs_core::sdfs::client::SdfsService;
use insane_sdfs_core::sdfs::types::types::{GetFileResult, ListFilesResult, SdfsFile};

pub struct TcpSdfsClient {}

#[async_trait::async_trait]
impl SdfsService for TcpSdfsClient {
    async fn list_files(&self, directory: &str) -> ListFilesResult {
        let exists = self.file_exists(directory).await.unwrap();

        if exists {
            // Directory exists, list files
            let mut entries = vec![];

            // Check if it's a directory
            if !tokio::fs::metadata(directory).await.unwrap().is_dir() {
                return ListFilesResult::FileNotDirectory;
            }

            // read the directory entries
            let mut read_dir = tokio::fs::read_dir(directory).await.unwrap();

            // Read directory entries
            loop {
                // Get the results
                let result = read_dir.next_entry().await.unwrap();

                match result {
                    // Process each entry
                    Some(entry) => {
                        // grab the file metadata
                        let result = self.get_file(&entry.path().to_string_lossy()).await;

                        // pushing the file entry
                        match result {
                            GetFileResult::Success(file) => entries.push(file),
                            _ => continue,
                        }
                    }
                    // break if no more entries
                    None => break,
                }
            }

            return ListFilesResult::Success(entries);
        } else {
            return ListFilesResult::DirectoryNotFound;
        }
    }

    async fn get_file(&self, path: &str) -> insane_sdfs_core::sdfs::types::types::GetFileResult {
        let exists = self.file_exists(path).await.unwrap();

        if exists {
            let metadata = tokio::fs::metadata(path).await.unwrap();
            let created_at = metadata.created().unwrap();
            let last_modified = metadata.modified().unwrap();

            if metadata.is_dir() {
                return GetFileResult::Success(SdfsFile::Directory {
                    name: path.to_string(),
                    created_at: chrono::DateTime::<chrono::Utc>::from(created_at),
                    last_modified: chrono::DateTime::<chrono::Utc>::from(last_modified),
                });
            } else {
                return GetFileResult::Success(SdfsFile::RegularFile {
                    name: path.to_string(),
                    size: metadata.len(),
                    created_at: chrono::DateTime::<chrono::Utc>::from(created_at),
                    last_modified: chrono::DateTime::<chrono::Utc>::from(last_modified),
                });
            }
        } else {
            return GetFileResult::FileNotFound;
        }
    }

    async fn file_exists(&self, path: &str) -> io::Result<bool> {
        return tokio::fs::try_exists(path).await;
    }

    async fn download_file(
        &self,
        _path: &str,
    ) -> insane_sdfs_core::sdfs::types::types::DownloadFileResult {
        unimplemented!()
    }

    async fn delete_file(
        &self,
        _path: &str,
    ) -> insane_sdfs_core::sdfs::types::types::DeleteFileResult {
        unimplemented!()
    }

    async fn upload_file(
        &self,
        _path: &str,
        _data: impl trpl::StreamExt<Item = Vec<u8>> + Send,
    ) -> io::Result<bool> {
        unimplemented!()
    }

    async fn move_file(
        &self,
        _src_path: &str,
        _dest_path: &str,
        _overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn rename_file(
        &self,
        _src_path: &str,
        _new_name: &str,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn copy_file(
        &self,
        _src_path: &str,
        _dest_path: &str,
        _overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::types::MoveFileResult {
        unimplemented!()
    }

    async fn create_directory(&self, path: &str, recursive: bool) -> io::Result<bool> {
        if recursive {
            return Ok(tokio::fs::create_dir_all(path).await.is_ok());
        } else {
            return Ok(tokio::fs::create_dir(path).await.is_ok());
        }
    }

    async fn ping(&self) -> bool {
        return true;
    }
}

#[cfg(test)]
mod tests {
    use std::fs;

    use super::*;

    #[tokio::test]
    async fn test_ping() {
        let client = TcpSdfsClient {};
        let result = client.ping().await;
        assert!(result);
    }

    #[tokio::test]
    async fn test_file_exists() {
        let client = TcpSdfsClient {};
        let path = "test_file_exists.txt";
        let contents = "Hello, Insane SDFS!";

        // Ensure the file does not exist
        let result = client.file_exists(path).await;
        assert!(result.is_ok());

        // Create the file and test existence again
        assert!(fs::write(path, contents).is_ok());
        let result = client.file_exists(path).await;
        assert!(result.is_ok());

        // Clean up
        assert!(fs::remove_file(path).is_ok());
        let result = client.file_exists(path).await;
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[tokio::test]
    async fn test_list_files() {
        let client = TcpSdfsClient {};
        let _result = client.list_files("/some/directory").await;
        // Add assertions here based on expected behavior
    }
}
