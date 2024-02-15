import React, { useState, useEffect } from 'react';
import {
  AppBar,
  Button,
  Card,
  CardContent,
  Container,
  CssBaseline,
  Toolbar,
  Typography,
  createTheme,
  ThemeProvider,
} from '@mui/material';
import LiveData from './LiveData'; // Adjust the import path based on your project structure

const theme = createTheme();

const Block = ({ block }) => {
  const [expanded, setExpanded] = useState(false);

  const toggleExpansion = () => {
    setExpanded(!expanded);
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h6">Block Hash: {block.HashPrev}</Typography>
        <Typography variant="subtitle1">Nonce: {block.Nonce}</Typography>
        <Typography variant="subtitle1">Sources:</Typography>
        {block.Sources ? (
          block.Sources.map((source, index) => (
            <div key={index}>
              <Typography variant="subtitle2">ID: {source.Id}</Typography>

              {source.Chunks.length > 0 ? (
                expanded && (
                  <div>
                    <Typography variant="subtitle2">Chunks:</Typography>
                    {source.Chunks.map((chunk, index) => (
                      <div key={index}>
                        <Typography variant="subtitle2">BytesProcessed: {chunk.BytesProcessed}</Typography>
                        <Typography variant="subtitle2">Hash: {chunk.Hash}</Typography>
                        <Typography variant="subtitle2">Signature: {chunk.Signature}</Typography>
                      </div>
                    ))}
                  </div>
                )
              ) : (
                <Typography variant="subtitle2">No chunk</Typography>
              )}
            </div>
          ))
        ) : (
          <Typography variant="subtitle2">No Sources available</Typography>
        )}
        <Button variant="text" color="primary" onClick={toggleExpansion}>
          {expanded ? 'View Less' : 'View More'}
        </Button>
      </CardContent>
    </Card>
  );
};

const App = () => {
  const [blocks, setBlocks] = useState([]);
  const [showBlocks, setShowBlocks] = useState(false);
  const [showLive, setLive] = useState(false);
  const getRecentBlocks = async () => {
    try {
      const response = await fetch('http://localhost:6963/getblocksrecent');
      const data = await response.json();
      setBlocks(data);
      setLive(false);
      setShowBlocks(true);
    } catch (error) {
      console.error('Error fetching recent blocks:', error);
    }
  };
  const [liveSources, setLiveSources] = useState([]);

  useEffect(() => {
    // Fetch the list of live sources
    const fetchLiveSources = async () => {
      try {
        const response = await fetch('http://localhost:6963/getlivesources');
        const data = await response.json();
        setLiveSources(data);
      } catch (error) {
        console.error('Error fetching live sources:', error);
      }
    };

    fetchLiveSources();
  }, []);
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6">Data verification</Typography>
          <Button color="inherit" onClick={getRecentBlocks}>
            Blocks
          </Button>
          <Button color="inherit" onClick={() => {
            setShowBlocks(false)
            setLive(true)
            
          }}>LiveData</Button>
          {/* Add other buttons as needed */}
        </Toolbar>
      </AppBar>
      <Container>
        {showBlocks && (
          <div style={{ display: 'flex', flexWrap: 'wrap' }}>
            {blocks.map((block, index) => (
              <Block key={index} block={block} />
            ))}
          </div>
        )}

        {/* Integration of LiveData component */}
        {showLive && (
        <LiveData liveSources={liveSources} />
        )}
      </Container>
    </ThemeProvider>
  );
};

export default App;
