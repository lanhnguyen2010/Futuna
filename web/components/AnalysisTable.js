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

  const colorFormatter = (cell) => {
    const value = cell.getValue();
    if (value.startsWith('ACCUMULATE')) {
      cell.getElement().style.color = 'green';
    } else if (value.startsWith('AVOID')) {
      cell.getElement().style.color = 'red';
    } else {
      cell.getElement().style.color = 'orange';
    }
    return value;
  };

  const columns = [
    { title: 'Ticker', field: 'ticker', hozAlign: 'left' },
    { title: 'Short', field: 'short_term', formatter: colorFormatter },
    { title: 'Long', field: 'long_term', formatter: colorFormatter },
    { title: 'Overall', field: 'overall', formatter: colorFormatter }
  ];

  return <ReactTabulator data={data} columns={columns} layout="fitData" />;
}
