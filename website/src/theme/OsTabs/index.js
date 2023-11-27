import React, { useState, useEffect } from 'react';
import OriginalTabs from '@theme/Tabs';
import { osName } from 'react-device-detect';

function OsTabs(props) {
  const [defaultValue, setDefaultValue] = useState(null);

  useEffect(() => {
    // Based on the OS or any other client-side condition, set the default value
    const tabs = ['iOS', 'Android', 'Mac OS', 'Windows', 'Linux']
    if (tabs.includes(osName)) {
      setDefaultValue(osName);
    } else {
      setDefaultValue('Windows');
    }
  }, []); // The empty dependency array ensures this runs once after component mount

  return (
    <>
      <OriginalTabs {...props} defaultValue={defaultValue}>
        {props.children}
      </OriginalTabs>
      {/* Uncomment the following line to debug the detected and selected values */}
      {/* <h2>detected={osName}, selected={defaultValue}</h2> */}
    </>
  );
}

export default OsTabs;
