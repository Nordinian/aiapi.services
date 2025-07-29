import React, { useEffect } from 'react';

const TokenQuery = () => {
  useEffect(() => {
    // Redirect to external token query service
    window.location.href = 'https://query.aiapi.services/';
  }, []);

  return (
    <div className="flex justify-center items-center h-screen">
      <div className="text-center">
        <p>正在跳转到令牌查询页面...</p>
        <p>Redirecting to token query page...</p>
      </div>
    </div>
  );
};

export default TokenQuery;