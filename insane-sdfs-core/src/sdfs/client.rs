pub mod client {
    use trpl::StreamExt;

    use crate::sdfs::types::types::{
        DeleteFileResult, DownloadFileResult, GetFileResult, ListFilesResult, MoveFileResult,
        SdfsConnectionConfig, SdfsResult,
    };

    pub trait SdfsService {
        // File querying methods
        fn list_files<'a>(&'a self, directory: &str) -> impl Future<Output = ListFilesResult> + 'a;

        fn get_file<'a>(&'a self, path: &str) -> impl Future<Output = GetFileResult> + 'a;

        fn file_exists<'a>(&'a self, path: &str) -> impl Future<Output = bool> + 'a;

        fn download_file<'a>(&'a self, path: &str)
        -> impl Future<Output = DownloadFileResult> + 'a;

        // File manipulation methods
        fn delete_file<'a>(&'a self, path: &str) -> impl Future<Output = DeleteFileResult> + 'a;

        fn upload_file<'a>(
            &'a self,
            path: &str,
            data: impl StreamExt<Item = Vec<u8>> + 'a,
        ) -> impl Future<Output = bool> + 'a;

        fn move_file<'a>(
            &'a self,
            src_path: &str,
            dest_path: &str,
            overwrite_existing: bool,
        ) -> impl Future<Output = MoveFileResult> + 'a;

        fn rename_file<'a>(
            &'a self,
            src_path: &str,
            new_name: &str,
        ) -> impl Future<Output = MoveFileResult> + 'a;

        fn copy_file<'a>(
            &'a self,
            src_path: &str,
            dest_path: &str,
            overwrite_existing: bool,
        ) -> impl Future<Output = MoveFileResult> + 'a;

        fn create_directory<'a>(
            &'a self,
            path: &str,
            recursive: bool,
        ) -> impl Future<Output = bool> + 'a;

        fn ping<'a>(&'a self) -> impl Future<Output = bool> + 'a;
    }

    pub trait SdfsServiceFactory<T: SdfsService> {
        fn connect<'a>(
            &'a self,
            config: &SdfsConnectionConfig,
        ) -> impl Future<Output = SdfsResult<T>> + 'a;
    }
}
