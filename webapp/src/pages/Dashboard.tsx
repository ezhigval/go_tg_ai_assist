import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'

export default function Dashboard() {
  const cards = [
    {
      title: 'Задачи',
      icon: '📝',
      path: '/todo',
      color: 'bg-blue-500',
    },
    {
      title: 'Финансы',
      icon: '💰',
      path: '/finance',
      color: 'bg-green-500',
    },
    {
      title: 'Кредиты',
      icon: '🏦',
      path: '/credits',
      color: 'bg-purple-500',
    },
  ]

  return (
    <div className="min-h-screen bg-secondary-bg p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="max-w-md mx-auto"
      >
        <h1 className="text-2xl font-bold mb-6 text-text">Главная</h1>

        <div className="space-y-3">
          {cards.map((card, index) => (
            <motion.div
              key={card.path}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Link
                to={card.path}
                className="block p-4 rounded-xl bg-bg shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="flex items-center gap-4">
                  <div className={`w-12 h-12 rounded-lg ${card.color} flex items-center justify-center text-2xl`}>
                    {card.icon}
                  </div>
                  <div className="flex-1">
                    <h2 className="text-lg font-semibold text-text">{card.title}</h2>
                  </div>
                  <div className="text-hint">→</div>
                </div>
              </Link>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </div>
  )
}

