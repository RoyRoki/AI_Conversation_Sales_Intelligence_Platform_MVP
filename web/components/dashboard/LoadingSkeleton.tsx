export default function LoadingSkeleton() {
  return (
    <div className="max-w-7xl mx-auto animate-pulse">
      {/* Header Skeleton */}
      <div className="mb-8">
        <div className="h-8 bg-gray-700 rounded w-64 mb-2"></div>
        <div className="h-4 bg-gray-700 rounded w-48"></div>
      </div>

      {/* Metrics Grid Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-gray-800 rounded-lg shadow-professional p-6 border border-gray-700">
            <div className="h-4 bg-gray-700 rounded w-32 mb-4"></div>
            <div className="h-10 bg-gray-700 rounded w-24"></div>
          </div>
        ))}
      </div>

      {/* Data Visualization Skeleton */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        {[1, 2].map((i) => (
          <div key={i} className="bg-gray-800 rounded-lg shadow-professional p-6 border border-gray-700">
            <div className="h-6 bg-gray-700 rounded w-40 mb-6"></div>
            <div className="space-y-3">
              {[1, 2, 3, 4].map((j) => (
                <div key={j} className="h-4 bg-gray-700 rounded"></div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Quick Actions Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div key={i} className="bg-gray-800 rounded-lg shadow-professional p-6 border border-gray-700">
            <div className="h-6 bg-gray-700 rounded w-40 mb-2"></div>
            <div className="h-4 bg-gray-700 rounded w-full"></div>
          </div>
        ))}
      </div>
    </div>
  );
}

