package com.featurevoting.adapters

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.core.content.ContextCompat
import androidx.recyclerview.widget.RecyclerView
import com.featurevoting.R
import com.featurevoting.databinding.ItemFeatureBinding
import com.featurevoting.models.Feature

class FeatureAdapter(
    private val features: List<Feature>,
    private val onVoteClick: (Feature, Boolean) -> Unit
) : RecyclerView.Adapter<FeatureAdapter.FeatureViewHolder>() {
    
    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): FeatureViewHolder {
        val binding = ItemFeatureBinding.inflate(
            LayoutInflater.from(parent.context),
            parent,
            false
        )
        return FeatureViewHolder(binding)
    }
    
    override fun onBindViewHolder(holder: FeatureViewHolder, position: Int) {
        holder.bind(features[position])
    }
    
    override fun getItemCount(): Int = features.size
    
    inner class FeatureViewHolder(
        private val binding: ItemFeatureBinding
    ) : RecyclerView.ViewHolder(binding.root) {
        
        fun bind(feature: Feature) {
            binding.apply {
                tvTitle.text = feature.title
                tvDescription.text = feature.description
                tvCreatedBy.text = root.context.getString(R.string.created_by, feature.createdByUser)
                tvVoteCount.text = root.context.getString(R.string.votes_count, feature.voteCount)
                
                // Update vote button
                if (feature.hasUserVoted) {
                    btnVote.text = root.context.getString(R.string.unvote)
                    btnVote.setBackgroundColor(
                        ContextCompat.getColor(root.context, R.color.voted_button)
                    )
                } else {
                    btnVote.text = root.context.getString(R.string.vote)
                    btnVote.setBackgroundColor(
                        ContextCompat.getColor(root.context, R.color.vote_button)
                    )
                }
                
                btnVote.setOnClickListener {
                    onVoteClick(feature, !feature.hasUserVoted)
                }
            }
        }
    }
}