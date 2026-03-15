use futures::io;
use libp2p::{Multiaddr, TransportError, core::transport::ListenerId, noise};

#[async_trait::async_trait]
pub trait NetworkFileSystem {
    async fn write_file(&mut self, path: String, data: Vec<u8>) -> Result<(), io::Error>;
    async fn read_file(&mut self, path: String) -> Result<Vec<u8>, io::Error>;
    async fn delete_file(&mut self, path: String) -> Result<(), io::Error>;
}

mod networking {
    use libp2p::{request_response, swarm::NetworkBehaviour};

    #[derive(Debug, NetworkBehaviour)]
    pub struct FilesystemBehavior {
        request_response: request_response::Behaviour<request_response::cbor::Codec>,
    }
}

pub struct NetworkClient {
    swarm: libp2p::Swarm<networking::FilesystemBehavior>,
}

impl NetworkClient {
    async fn new() -> Self {
        let mut swarm = libp2p::SwarmBuilder::with_new_identity()
            .with_tokio()
            .with_tcp(
                tcp::Config::default(),
                noise::Config::new,
                yamux::Config::default,
            )?
            .with_quic()
            .with_behavior(|key| {});

        return Self { swarm };
    }

    async fn connect(&mut self, addr: Multiaddr) -> Result<ListenerId, TransportError<io::Error>> {}
}

pub struct NetworkServer {}

impl NetworkServer {
    async fn listen(&mut self, addr: Multiaddr) -> Result<ListenerId, TransportError<io::Error>> {}
}
