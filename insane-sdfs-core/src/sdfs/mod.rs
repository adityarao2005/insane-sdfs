pub mod sdfs {
    use chrono::{DateTime, Utc};
    use trpl::StreamExt;

    pub struct SdfsConnectionConfig {
        pub address: String,
        pub port: u16,
        pub token: String,
    }

    pub struct SdfsFile {
        pub name: String,
        pub size: u64,
        pub created_at: DateTime<Utc>,
        pub last_modified: DateTime<Utc>,
    }

    pub trait SdfsClientFactory<T: SdfsClient> {
        fn connect<'a>(
            &'a self,
            config: &SdfsConnectionConfig,
        ) -> impl Future<Output = Option<T>> + 'a;
    }

    pub trait SdfsClient {
        // File querying methods
        fn list_files<'a>(&'a self, directory: &str) -> impl Future<Output = Vec<SdfsFile>> + 'a;

        fn get_file<'a>(&'a self, path: &str) -> impl Future<Output = Option<SdfsFile>> + 'a;

        fn file_exists<'a>(&'a self, path: &str) -> impl Future<Output = bool> + 'a;

        fn download_file<'a>(
            &'a self,
            path: &str,
        ) -> impl Future<Output = Option<impl StreamExt<Item = Vec<u8>> + 'a>> + 'a;

        // File manipulation methods
        fn delete_file<'a>(&'a self, path: &str) -> impl Future<Output = bool> + 'a;

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
        ) -> impl Future<Output = bool> + 'a;

        fn rename_file<'a>(
            &'a self,
            src_path: &str,
            new_name: &str,
        ) -> impl Future<Output = bool> + 'a;

        fn copy_file<'a>(
            &'a self,
            src_path: &str,
            dest_path: &str,
            overwrite_existing: bool,
        ) -> impl Future<Output = bool> + 'a;

        fn create_directory<'a>(
            &'a self,
            path: &str,
            recursive: bool,
        ) -> impl Future<Output = bool> + 'a;
    }
}
