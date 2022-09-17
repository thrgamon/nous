import dynamic from 'next/dynamic'
import type { NextPage } from 'next'

const Drawing: NextPage = () => {
  const Scali = dynamic(
    () => import('../components/scali'),
    { ssr: false }
  )
  return (
    <div className="w-screen h-screen">
      <Scali />
    </div>
  );
}

export default Drawing
