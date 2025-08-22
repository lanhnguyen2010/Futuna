import React, { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';

const ReactTabulator = dynamic(() => import('react-tabulator'), { ssr: false });

export default function AnalysisTable() {
  const [data, setData] = useState([]);
  const api = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  useEffect(() => {
    fetch(`${api}/api/analysis`)
      .then((res) => res.json())
      .then(setData)
      .catch(console.error);
  }, [api]);

  const columns = [
    { title: 'Ticker', field: 'ticker', hozAlign: 'left' },
    { title: 'Short', field: 'short_term.rating' },
    { title: 'Long', field: 'long_term.rating' }
  ];

  return <ReactTabulator data={data} columns={columns} layout="fitData" />;
}
