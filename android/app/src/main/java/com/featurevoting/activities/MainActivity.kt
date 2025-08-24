package com.featurevoting.activities

import android.content.Intent
import android.os.Bundle
import android.view.Menu
import android.view.MenuItem
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import androidx.recyclerview.widget.LinearLayoutManager
import com.featurevoting.R
import com.featurevoting.adapters.FeatureAdapter
import com.featurevoting.api.ApiClient
import com.featurevoting.databinding.ActivityMainBinding
import com.featurevoting.models.Feature
import com.featurevoting.utils.PreferenceManager
import com.featurevoting.utils.showToast
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {
    
    private lateinit var binding: ActivityMainBinding
    private lateinit var preferenceManager: PreferenceManager
    private lateinit var featureAdapter: FeatureAdapter
    private val features = mutableListOf<Feature>()
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)
        
        preferenceManager = PreferenceManager(this)
        
        // Check if logged in
        if (!preferenceManager.isLoggedIn()) {
            startLoginActivity()
            return
        }
        
        setupViews()
        loadFeatures()
    }
    
    private fun setupViews() {
        setSupportActionBar(binding.toolbar)
        
        // Setup RecyclerView
        featureAdapter = FeatureAdapter(features) { feature, isVoting ->
            if (isVoting) {
                voteForFeature(feature)
            } else {
                removeVoteFromFeature(feature)
            }
        }
        
        binding.rvFeatures.apply {
            layoutManager = LinearLayoutManager(this@MainActivity)
            adapter = featureAdapter
        }
        
        // Setup SwipeRefresh
        binding.swipeRefresh.setOnRefreshListener {
            loadFeatures()
        }
        
        // Setup FAB
        binding.fabCreateFeature.setOnClickListener {
            startActivity(Intent(this, CreateFeatureActivity::class.java))
        }
    }
    
    override fun onResume() {
        super.onResume()
        if (preferenceManager.isLoggedIn()) {
            loadFeatures()
        }
    }
    
    override fun onCreateOptionsMenu(menu: Menu?): Boolean {
        menuInflater.inflate(R.menu.main_menu, menu)
        return true
    }
    
    override fun onOptionsItemSelected(item: MenuItem): Boolean {
        return when (item.itemId) {
            R.id.action_logout -> {
                logout()
                true
            }
            else -> super.onOptionsItemSelected(item)
        }
    }
    
    private fun loadFeatures() {
        showLoading(true)
        
        lifecycleScope.launch {
            try {
                val response = ApiClient.getApiService().getFeatures()
                
                if (response.isSuccessful && response.body() != null) {
                    val featuresResponse = response.body()!!
                    features.clear()
                    features.addAll(featuresResponse.features)
                    featureAdapter.notifyDataSetChanged()
                    
                    binding.tvEmpty.visibility = 
                        if (features.isEmpty()) View.VISIBLE else View.GONE
                } else {
                    showToast("Failed to load features")
                }
            } catch (e: Exception) {
                showToast("Network error. Please check your connection.")
            } finally {
                showLoading(false)
            }
        }
    }
    
    private fun voteForFeature(feature: Feature) {
        lifecycleScope.launch {
            try {
                val response = ApiClient.getApiService().voteForFeature(feature.id)
                
                if (response.isSuccessful && response.body() != null) {
                    val voteResponse = response.body()!!
                    
                    // Update the feature in the list
                    val index = features.indexOfFirst { it.id == feature.id }
                    if (index != -1) {
                        val updatedFeature = features[index].copy(
                            voteCount = voteResponse.voteCount,
                            hasUserVoted = voteResponse.hasVoted
                        )
                        features[index] = updatedFeature
                        featureAdapter.notifyItemChanged(index)
                    }
                    
                    showToast("Vote added!")
                } else {
                    showToast("Failed to vote")
                }
            } catch (e: Exception) {
                showToast("Network error")
            }
        }
    }
    
    private fun removeVoteFromFeature(feature: Feature) {
        lifecycleScope.launch {
            try {
                val response = ApiClient.getApiService().removeVoteFromFeature(feature.id)
                
                if (response.isSuccessful && response.body() != null) {
                    val voteResponse = response.body()!!
                    
                    // Update the feature in the list
                    val index = features.indexOfFirst { it.id == feature.id }
                    if (index != -1) {
                        val updatedFeature = features[index].copy(
                            voteCount = voteResponse.voteCount,
                            hasUserVoted = voteResponse.hasVoted
                        )
                        features[index] = updatedFeature
                        featureAdapter.notifyItemChanged(index)
                    }
                    
                    showToast("Vote removed!")
                } else {
                    showToast("Failed to remove vote")
                }
            } catch (e: Exception) {
                showToast("Network error")
            }
        }
    }
    
    private fun showLoading(show: Boolean) {
        binding.swipeRefresh.isRefreshing = show
        if (!show) {
            binding.progressBar.visibility = View.GONE
        }
    }
    
    private fun logout() {
        preferenceManager.clearUserData()
        ApiClient.clearApiService()
        startLoginActivity()
    }
    
    private fun startLoginActivity() {
        startActivity(Intent(this, LoginActivity::class.java))
        finish()
    }
}