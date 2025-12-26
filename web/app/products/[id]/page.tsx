'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { apiClient } from '@/lib/api';
import type { Product } from '@/types';
import CustomerChat from '@/components/CustomerChat';
import FloatingChatButton from '@/components/FloatingChatButton';

export default function ProductDetailPage() {
  const params = useParams();
  const productId = params.id as string;
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (productId) {
      loadProduct();
    }
  }, [productId]);

  const loadProduct = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getProduct(productId, 'OMX26');
      setProduct(data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load product');
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
      <div className="min-h-screen bg-gray-900 text-white">
        <div className="max-w-7xl mx-auto px-4 py-8">
          <div className="text-center text-gray-400">Loading product...</div>
        </div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="min-h-screen bg-gray-900 text-white">
        <div className="max-w-7xl mx-auto px-4 py-8">
          <div className="bg-red-900/20 border border-red-500 text-red-400 px-4 py-3 rounded">
            {error || 'Product not found'}
          </div>
          <Link href="/customer/products" className="mt-4 inline-block text-blue-400 hover:text-blue-300">
            ← Back to Products
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <div className="max-w-7xl mx-auto px-4 py-8">
        <Link href="/customer/products" className="text-blue-400 hover:text-blue-300 mb-4 inline-block">
          ← Back to Products
        </Link>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2">
            <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6 mb-6">
              <h1 className="text-4xl font-bold mb-4 text-blue-400">{product.name}</h1>
              <p className="text-gray-300 mb-6 text-lg">{product.description}</p>

              <div className="mb-6">
                <div className="text-4xl font-bold text-white mb-2">
                  {formatPrice(product.price, product.price_currency)}
                </div>
                {product.price > 0 && (
                  <span className="text-gray-400">per month</span>
                )}
              </div>

              <div className="mb-6">
                <h2 className="text-xl font-semibold mb-3">Target Audience</h2>
                <p className="text-gray-300">{product.target_audience}</p>
              </div>

              <div className="mb-6">
                <h2 className="text-xl font-semibold mb-3">Features</h2>
                <ul className="list-disc list-inside text-gray-300 space-y-2">
                  {product.features.map((feature, idx) => (
                    <li key={idx}>{feature}</li>
                  ))}
                </ul>
              </div>

              {product.limitations.length > 0 && (
                <div className="mb-6">
                  <h2 className="text-xl font-semibold mb-3">Limitations</h2>
                  <ul className="list-disc list-inside text-gray-300 space-y-2">
                    {product.limitations.map((limitation, idx) => (
                      <li key={idx}>{limitation}</li>
                    ))}
                  </ul>
                </div>
              )}

              {product.common_questions.length > 0 && (
                <div>
                  <h2 className="text-xl font-semibold mb-3">Common Questions</h2>
                  <ul className="list-disc list-inside text-gray-300 space-y-2">
                    {product.common_questions.map((question, idx) => (
                      <li key={idx}>{question}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </div>

          <div className="lg:col-span-1">
            <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-6 sticky top-4">
              <h2 className="text-2xl font-bold mb-4">Chat with Sales</h2>
              <p className="text-gray-400 mb-4 text-sm">
                Have questions? Our AI-powered sales team is here to help!
              </p>
              <CustomerChat productId={productId} />
            </div>
          </div>
        </div>
      </div>
      <FloatingChatButton />
    </div>
  );
}

