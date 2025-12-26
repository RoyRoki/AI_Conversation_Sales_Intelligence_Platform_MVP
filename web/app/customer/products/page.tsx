'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api';
import type { Product } from '@/types';
import GlobalChat from '@/components/GlobalChat';

export default function CustomerProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadProducts();
  }, []);

  const loadProducts = async () => {
    try {
      setLoading(true);
      const data = await apiClient.listProducts('OMX26');
      setProducts(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load products');
    } finally {
      setLoading(false);
    }
  };

  const formatPrice = (price: number, currency: string) => {
    if (price === 0) return 'Contact Sales';
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: currency || 'INR',
    }).format(price);
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="text-center text-gray-400">Loading products...</div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-white mb-2">All Products</h1>
        <p className="text-gray-400">Browse our complete product catalog</p>
      </div>

      {error && (
        <div className="mb-4 bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Products List - Left Side */}
        <div className="lg:col-span-2">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {products.map((product) => (
              <div
                key={product.id}
                className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6 hover:border-blue-500 transition-colors"
              >
                <h2 className="text-2xl font-bold mb-2 text-blue-400">{product.name}</h2>
                <p className="text-gray-300 mb-4 text-sm">{product.description}</p>
                
                <div className="mb-4">
                  <div className="text-3xl font-bold text-white mb-1">
                    {formatPrice(product.price, product.price_currency)}
                  </div>
                  {product.price > 0 && (
                    <span className="text-gray-400 text-sm">per month</span>
                  )}
                </div>

                <div className="mb-4">
                  <p className="text-sm text-gray-400 mb-2">Target Audience:</p>
                  <p className="text-sm text-gray-300">{product.target_audience}</p>
                </div>

                <div className="mb-4">
                  <p className="text-sm text-gray-400 mb-2">Key Features:</p>
                  <ul className="list-disc list-inside text-sm text-gray-300 space-y-1">
                    {product.features.slice(0, 3).map((feature, idx) => (
                      <li key={idx}>{feature}</li>
                    ))}
                  </ul>
                </div>

                <div className="flex space-x-2">
                  <Link
                    href={`/customer/chat?product=${product.id}`}
                    className="flex-1 text-center bg-green-600 hover:bg-green-700 text-white font-medium py-2 px-4 rounded transition-colors text-sm"
                  >
                    Chat about this
                  </Link>
                  <Link
                    href={`/products/${product.id}`}
                    className="flex-1 text-center bg-gray-600 hover:bg-gray-700 text-white font-medium py-2 px-4 rounded transition-colors text-sm"
                  >
                    View Details
                  </Link>
                </div>
              </div>
            ))}
          </div>

          {products.length === 0 && !loading && (
            <div className="text-center py-12 text-gray-400">No products available</div>
          )}
        </div>

        {/* Global Chat - Right Side */}
        <div className="lg:col-span-1">
          <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6 sticky top-4">
            <h2 className="text-2xl font-bold mb-4">Chat with Sales</h2>
            <p className="text-gray-400 mb-4 text-sm">
              Have questions? Our AI-powered sales team is here to help!
            </p>
            <div className="h-[600px]">
              <GlobalChat />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

