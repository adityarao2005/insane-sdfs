use std::io;

use insane_sdfs_core::sdfs::client::SdfsService;
use insane_sdfs_core::sdfs::types::{GetFileResult, ListFilesResult, SdfsFile};

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

    async fn get_file(&self, path: &str) -> insane_sdfs_core::sdfs::types::GetFileResult {
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
    ) -> insane_sdfs_core::sdfs::types::DownloadFileResult {
        unimplemented!()
    }

    async fn delete_file(&self, path: &str) -> insane_sdfs_core::sdfs::types::DeleteFileResult {
        let file = self.get_file(path).await;

        match file {
            GetFileResult::Success(sdfs_file) => match sdfs_file {
                SdfsFile::RegularFile { .. } => {
                    let _ = tokio::fs::remove_file(path).await.unwrap();
                    return insane_sdfs_core::sdfs::types::DeleteFileResult::Success;
                }
                SdfsFile::Directory { .. } => {
                    let result = self.list_files(path).await;

                    if let ListFilesResult::Success(files) = result {
                        if files.is_empty() {
                            let _ = tokio::fs::remove_dir(path).await.unwrap();
                            return insane_sdfs_core::sdfs::types::DeleteFileResult::Success;
                        } else {
                            return insane_sdfs_core::sdfs::types::DeleteFileResult::DirectoryNotEmpty;
                        }
                    }

                    return insane_sdfs_core::sdfs::types::DeleteFileResult::FileNotFound;
                }
            },
            _ => {
                return insane_sdfs_core::sdfs::types::DeleteFileResult::FileNotFound;
            }
        }
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
        src_path: &str,
        dest_path: &str,
        overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::MoveFileResult {
        if !self.file_exists(src_path).await.unwrap_or(false) {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
        }

        let dest_exists = self.file_exists(dest_path).await.unwrap_or(false);

        if dest_exists && !overwrite_existing {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileAlreadyExists;
        } else if dest_exists {
            // get rid of the destination file
            match self.get_file(dest_path).await {
                GetFileResult::Success(value) => match value {
                    SdfsFile::RegularFile { .. } => {
                        if tokio::fs::remove_file(dest_path).await.is_err() {
                            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                        }
                    }
                    SdfsFile::Directory { .. } => {
                        if tokio::fs::remove_dir_all(dest_path).await.is_err() {
                            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                        }
                    }
                },
                _ => return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound,
            }
        } else {
            // Check parent directory exists
            if let Some(parent) = std::path::Path::new(dest_path)
                .parent()
                .and_then(|p| p.to_str())
            {
                // Empty string means current directory, which always exists
                if !parent.is_empty() && !self.file_exists(parent).await.unwrap_or(false) {
                    return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                }
            }
        }

        // rename the source to destination
        if tokio::fs::rename(src_path, dest_path).await.is_err() {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
        }

        insane_sdfs_core::sdfs::types::MoveFileResult::Success
    }

    async fn rename_file(
        &self,
        src_path: &str,
        new_name: &str,
    ) -> insane_sdfs_core::sdfs::types::MoveFileResult {
        return self.move_file(src_path, new_name, false).await;
    }

    async fn copy_file(
        &self,
        src_path: &str,
        dest_path: &str,
        overwrite_existing: bool,
    ) -> insane_sdfs_core::sdfs::types::MoveFileResult {
        if !self.file_exists(src_path).await.unwrap_or(false) {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
        }

        let dest_exists = self.file_exists(dest_path).await.unwrap_or(false);

        if dest_exists && !overwrite_existing {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileAlreadyExists;
        } else if dest_exists {
            // get rid of the destination file
            match self.get_file(dest_path).await {
                GetFileResult::Success(value) => match value {
                    SdfsFile::RegularFile { .. } => {
                        if tokio::fs::remove_file(dest_path).await.is_err() {
                            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                        }
                    }
                    SdfsFile::Directory { .. } => {
                        if tokio::fs::remove_dir_all(dest_path).await.is_err() {
                            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                        }
                    }
                },
                _ => return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound,
            }
        } else {
            // Check parent directory exists
            if let Some(parent) = std::path::Path::new(dest_path)
                .parent()
                .and_then(|p| p.to_str())
            {
                // Empty string means current directory, which always exists
                if !parent.is_empty() && !self.file_exists(parent).await.unwrap_or(false) {
                    return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
                }
            }
        }

        // copy the source to destination
        if tokio::fs::copy(src_path, dest_path).await.is_err() {
            return insane_sdfs_core::sdfs::types::MoveFileResult::FileNotFound;
        }

        insane_sdfs_core::sdfs::types::MoveFileResult::Success
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
    async fn test_list_files() {
        let client = TcpSdfsClient {};

        let directory = "some";
        let result = client.list_files(directory).await;
        assert!(matches!(result, ListFilesResult::DirectoryNotFound));

        fs::create_dir(directory).unwrap();

        let result = client.list_files(directory).await;
        assert!(matches!(result, ListFilesResult::Success(_)));

        // test empty directory
        if let ListFilesResult::Success(files) = result {
            assert!(files.is_empty());
        }

        let contents = "Hello, Insane SDFS!";
        fs::write(format!("{}/file1.txt", directory), contents).unwrap();

        // test listing with one file
        let result = client.list_files(directory).await;
        assert!(matches!(result, ListFilesResult::Success(_)));
        if let ListFilesResult::Success(files) = result {
            assert_eq!(files.len(), 1);
            match &files[0] {
                SdfsFile::RegularFile { name, size, .. } => {
                    assert_eq!(name, "some/file1.txt");
                    assert_eq!(*size, contents.len() as u64);
                }
                _ => panic!("Expected RegularFile"),
            }
        }

        fs::remove_dir_all(directory).unwrap();
    }

    #[tokio::test]
    async fn test_get_file() {
        let client = TcpSdfsClient {};

        let content = "Hello, Insane SDFS!";
        let file = "test_file.txt";
        let directory = "test_dir";
        fs::write(file, content).unwrap();
        fs::create_dir(directory).unwrap();

        let result = client.get_file(file).await;
        assert!(matches!(result, GetFileResult::Success(_)));
        if let GetFileResult::Success(sdfs_file) = result {
            match sdfs_file {
                SdfsFile::RegularFile { name, size, .. } => {
                    assert_eq!(name, file);
                    assert_eq!(size, content.len() as u64);
                }
                _ => panic!("Expected RegularFile"),
            }
        }

        let result = client.get_file(directory).await;
        assert!(matches!(result, GetFileResult::Success(_)));
        if let GetFileResult::Success(sdfs_file) = result {
            match sdfs_file {
                SdfsFile::Directory { name, .. } => {
                    assert_eq!(name, directory);
                }
                _ => panic!("Expected Directory"),
            }
        }

        fs::remove_file(file).unwrap();
        fs::remove_dir_all(directory).unwrap();
    }

    #[tokio::test]
    async fn test_file_exists() {
        let client = TcpSdfsClient {};
        let path = "test_file_exists.txt";
        let contents = "Hello, Insane SDFS!";

        // Ensure the file does not exist
        let result = client.file_exists(path).await;
        assert!(result.is_ok());
        assert!(!result.unwrap());

        // Create the file and test existence again
        assert!(fs::write(path, contents).is_ok());
        let result = client.file_exists(path).await;
        assert!(result.is_ok());
        assert!(result.unwrap());

        // Clean up
        assert!(fs::remove_file(path).is_ok());
        let result = client.file_exists(path).await;
        assert!(result.is_ok());
        assert!(!result.unwrap());
    }

    #[tokio::test]
    async fn test_create_directory() {
        let client = TcpSdfsClient {};
        let path = "test_create_directory";
        let recursive_path = "test_create_directory/nested/dir";

        // Ensure the dirs does not exist
        let result = fs::exists(path);
        assert!(result.is_ok());
        assert!(!result.unwrap());
        let result = fs::exists(recursive_path);
        assert!(result.is_ok());
        assert!(!result.unwrap());

        // Create the dirs and test existence again
        assert!(client.create_directory(path, false).await.is_ok());
        let result = fs::exists(path);
        assert!(result.is_ok());
        assert!(result.unwrap());

        assert!(client.create_directory(recursive_path, true).await.is_ok());
        let result = fs::exists(recursive_path);
        assert!(result.is_ok());
        assert!(result.unwrap());

        // Clean up
        assert!(fs::remove_dir_all(path).is_ok());
    }

    #[tokio::test]
    async fn test_delete_file() {
        let client = TcpSdfsClient {};
        let file_path = "test_delete_file.txt";
        let dir_path = "test_delete_dir";
        let non_empty_dir_path = "test_delete_non_empty_dir";

        // Test deleting a regular file
        fs::write(file_path, "Hello, Insane SDFS!").unwrap();
        let result = client.delete_file(file_path).await;
        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::DeleteFileResult::Success
        ));
        assert!(!fs::exists(file_path).unwrap());

        // Test deleting an empty directory
        fs::create_dir(dir_path).unwrap();
        let result = client.delete_file(dir_path).await;
        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::DeleteFileResult::Success
        ));
        assert!(!fs::exists(dir_path).unwrap());

        // Test deleting a non-empty directory
        fs::create_dir(non_empty_dir_path).unwrap();
        fs::write(format!("{}/file.txt", non_empty_dir_path), "Content").unwrap();
        let result = client.delete_file(non_empty_dir_path).await;
        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::DeleteFileResult::DirectoryNotEmpty
        ));
        assert!(fs::exists(non_empty_dir_path).unwrap());

        // Clean up
        fs::remove_dir_all(non_empty_dir_path).unwrap();
    }

    #[tokio::test]
    async fn test_move_file() {
        let client = TcpSdfsClient {};
        let src_path = "test_move_src.txt";
        let dest_path = "test_move_dest.txt";

        // Create source file
        fs::write(src_path, "Hello, Insane SDFS!").unwrap();

        // Move file to destination
        let result = client.move_file(src_path, dest_path, false).await;

        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::MoveFileResult::Success
        ));
        assert!(!fs::exists(src_path).unwrap());
        assert!(fs::exists(dest_path).unwrap());

        // Clean up
        fs::remove_file(dest_path).unwrap();
    }

    #[tokio::test]
    async fn test_rename_file() {
        let client = TcpSdfsClient {};
        let src_path = "test_rename_src.txt";
        let new_name = "test_rename_new.txt";

        // Create source file
        fs::write(src_path, "Hello, Insane SDFS!").unwrap();
        // Rename file
        let result = client.rename_file(src_path, new_name).await;
        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::MoveFileResult::Success
        ));
        assert!(!fs::exists(src_path).unwrap());
        assert!(fs::exists(new_name).unwrap());

        // Clean up
        fs::remove_file(new_name).unwrap();
    }

    #[tokio::test]
    async fn test_copy_file() {
        let client = TcpSdfsClient {};
        let src_path = "test_copy_src.txt";
        let dest_path = "test_copy_dest.txt";

        // Create source file
        fs::write(src_path, "Hello, Insane SDFS!").unwrap();
        // Copy file to destination
        let result = client.copy_file(src_path, dest_path, false).await;
        assert!(matches!(
            result,
            insane_sdfs_core::sdfs::types::MoveFileResult::Success
        ));
        assert!(fs::exists(src_path).unwrap());
        assert!(fs::exists(dest_path).unwrap());
        let src_content = fs::read_to_string(src_path).unwrap();
        let dest_content = fs::read_to_string(dest_path).unwrap();
        assert_eq!(src_content, dest_content);

        // Clean up
        fs::remove_file(src_path).unwrap();
        fs::remove_file(dest_path).unwrap();
    }

    #[tokio::test]
    async fn test_ping() {
        let client = TcpSdfsClient {};
        let result = client.ping().await;
        assert!(result);
    }
}
