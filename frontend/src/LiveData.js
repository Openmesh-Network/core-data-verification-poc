import React, { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Typography,
} from '@mui/material';

const LiveData = ({ liveSources }) => {
  const [selectedSource, setSelectedSource] = useState('');
  const [liveData, setLiveData] = useState([]);

  useEffect(() => {
    // Fetch live data when selectedSource changes
    const fetchData = async () => {
      if (selectedSource) {
        try {
          const response = await fetch(`http://localhost:6963/getdata/${selectedSource}`);
          const data = await response.json();
          setLiveData(data);
        } catch (error) {
          console.error('Error fetching live data:', error);
        }
      }
    };

    fetchData();
  }, [selectedSource]);

  const handleSourceChange = (event) => {
    setSelectedSource(event.target.value);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
      <FormControl style={{ marginBottom: '20px' }}>
        <InputLabel id="live-source-label">Select Live Source</InputLabel>
        <Select
          labelId="live-source-label"
          id="live-source-select"
          value={selectedSource}
          onChange={handleSourceChange}
        >
          {liveSources.map((source, index) => (
            <MenuItem key={index} value={source.Name}>
              {source.Name}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      <div>
        {liveData.map((event, index) => (
          <Card key={index} style={{ marginBottom: '20px' }}>
            <CardContent>
              <Typography variant="h6">Timestamp: {event.Timestamp}</Typography>
              <Typography variant="subtitle1">Event: {event.Event}</Typography>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
};

export default LiveData;
