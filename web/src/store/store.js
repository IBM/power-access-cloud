import { configureStore } from '@reduxjs/toolkit'
import rootReducer from './reducer';
import storage from 'redux-persist/lib/storage';
import { persistReducer,persistStore } from 'redux-persist';
import { logger } from 'redux-logger';


// for Redux Persist
const persistConfig = {
  key: 'pacui-root',
  storage: storage
};

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer : persistedReducer,
  // redux-thunk v3 removed the default export — thunk is already included by
  // configureStore via RTK, so just append logger to the default middleware.
  middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(logger)
});

export const persistor = persistStore(store);

